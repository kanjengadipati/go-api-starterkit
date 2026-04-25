package auth

import (
	"log"
	"time"

	"go-api-starterkit/internal/modules/audit"
	tokenModule "go-api-starterkit/internal/modules/token"
	userModule "go-api-starterkit/internal/modules/user"
	"go-api-starterkit/internal/services"
	"go-api-starterkit/internal/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (s *authService) Register(user *userModule.User, password string) error {
	hashedPassword, err := services.HashPassword(password)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	user.Role = "user"
	user.IsVerified = false

	verificationToken := uuid.NewString()
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		txUserRepo := s.UserRepo.WithTx(tx)
		txEmailRepo := s.EmailVerificationRepo.WithTx(tx)

		if err := txUserRepo.Create(user); err != nil {
			return err
		}

		verificationRecord := &tokenModule.EmailVerificationToken{
			UserID:    user.ID,
			Token:     utils.HashToken(verificationToken),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			CreatedAt: time.Now(),
		}

		return txEmailRepo.Create(verificationRecord)
	})
	if err != nil {
		return err
	}

	if sendErr := s.EmailSvc.SendVerificationEmail(user.Email, verificationToken); sendErr != nil {
		log.Printf("verification email send failed for %s: %v", user.Email, sendErr)
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &user.ID,
		Action:      "register",
		Resource:    "user",
		ResourceID:  &user.ID,
		Status:      "success",
		Description: "user registered",
	})

	return nil
}

func (s *authService) Login(email, password, deviceID, userAgent, ipAddress string) (*AuthTokens, error) {
	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		s.AuditSvc.SafeRecord(audit.RecordInput{
			Action:      "login",
			Resource:    "auth",
			Status:      "failed",
			Description: "invalid credentials",
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
		})
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.AuditSvc.SafeRecord(audit.RecordInput{
			ActorUserID: &user.ID,
			Action:      "login",
			Resource:    "auth",
			ResourceID:  &user.ID,
			Status:      "failed",
			Description: "invalid credentials",
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
		})
		return nil, ErrInvalidCredentials
	}

	if !user.IsVerified {
		s.AuditSvc.SafeRecord(audit.RecordInput{
			ActorUserID: &user.ID,
			Action:      "login",
			Resource:    "auth",
			ResourceID:  &user.ID,
			Status:      "denied",
			Description: "email not verified",
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
		})
		return nil, ErrEmailNotVerified
	}

	tokens, err := s.issueTokens(user.ID, user.Role, user.AccessTokenVersion, deviceID, userAgent, ipAddress)
	if err != nil {
		return nil, err
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &user.ID,
		Action:      "login",
		Resource:    "auth",
		ResourceID:  &user.ID,
		Status:      "success",
		Description: "user logged in",
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})

	return tokens, nil
}

func (s *authService) GetProfile(userID uint) (*userModule.User, error) {
	return s.UserRepo.FindByID(userID)
}

func (s *authService) issueTokens(userID uint, role string, accessTokenVersion uint, deviceID, userAgent, ipAddress string) (*AuthTokens, error) {
	accessToken, err := s.JWT.GenerateToken(userID, role, 10*time.Minute, TokenAccess, accessTokenVersion)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.JWT.GenerateToken(userID, role, 7*24*time.Hour, TokenRefresh, accessTokenVersion)
	if err != nil {
		return nil, err
	}

	tokenHash := utils.HashToken(refreshToken)
	refreshTokenModel := &tokenModule.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		DeviceID:  deviceID,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		ExpiredAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.RefreshTokenRepo.Save(refreshTokenModel); err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
