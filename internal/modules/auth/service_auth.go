package auth

import (
	"context"
	"log"
	"time"

	"pleco-api/internal/modules/audit"
	tokenModule "pleco-api/internal/modules/token"
	userModule "pleco-api/internal/modules/user"
	"pleco-api/internal/services"
	"pleco-api/internal/utils"

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
	user.EmailVerified = false
	now := time.Now()
	user.PasswordUpdatedAt = now
	user.LastPasswordChange = &now

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
	ctx := context.Background()
	attemptKey := "login_attempts:" + email

	type loginAttempt struct {
		Count        int       `json:"count"`
		LockoutUntil time.Time `json:"lockout_until"`
	}

	var attempt loginAttempt
	if s.Cache != nil {
		ok, _ := s.Cache.GetJSON(ctx, attemptKey, &attempt)
		if ok && time.Now().Before(attempt.LockoutUntil) {
			s.AuditSvc.SafeRecord(audit.RecordInput{
				Action:      "login",
				Resource:    "auth",
				Status:      "failed",
				Description: "account locked due to too many failed attempts",
				IPAddress:   ipAddress,
				UserAgent:   userAgent,
			})
			return nil, ErrAccountLocked
		}
	}

	recordFailedAttempt := func(userID *uint) {
		if s.Cache == nil {
			return
		}
		attempt.Count++
		var lockoutDuration time.Duration
		switch {
		case attempt.Count < 4:
			lockoutDuration = 0
		case attempt.Count == 4:
			lockoutDuration = 1 * time.Minute
		case attempt.Count == 5:
			lockoutDuration = 5 * time.Minute
		case attempt.Count == 6:
			lockoutDuration = 15 * time.Minute
		default:
			lockoutDuration = 30 * time.Minute
		}

		ttl := lockoutDuration
		if ttl == 0 {
			ttl = 5 * time.Minute
		} else {
			attempt.LockoutUntil = time.Now().Add(lockoutDuration)
			ttl = lockoutDuration + 5*time.Minute
		}

		_ = s.Cache.SetJSON(ctx, attemptKey, attempt, ttl)

		s.AuditSvc.SafeRecord(audit.RecordInput{
			ActorUserID: userID,
			Action:      "login",
			Resource:    "auth",
			ResourceID:  userID,
			Status:      "failed",
			Description: "invalid credentials",
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
		})
	}

	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		recordFailedAttempt(nil)
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		recordFailedAttempt(&user.ID)
		return nil, ErrInvalidCredentials
	}

	if s.Cache != nil {
		_ = s.Cache.Delete(ctx, attemptKey)
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

	if err := s.UserRepo.UpdateLastLogin(user.ID, time.Now()); err != nil {
		return nil, err
	}
	s.invalidateUserCache(user.ID)

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
	tokens, _, err := s.issueTokensWithFamily(userID, role, accessTokenVersion, deviceID, userAgent, ipAddress, "", nil, true)
	return tokens, err
}

func (s *authService) issueTokensWithFamily(
	userID uint,
	role string,
	accessTokenVersion uint,
	deviceID, userAgent, ipAddress string,
	familyID string,
	rotatedFromTokenID *uint,
	replaceExistingDevice bool,
) (*AuthTokens, *tokenModule.RefreshToken, error) {
	tokens, refreshTokenModel, err := s.buildTokenPair(userID, role, accessTokenVersion, deviceID, userAgent, ipAddress, familyID, rotatedFromTokenID)
	if err != nil {
		return nil, nil, err
	}

	if err := s.runUserRefreshTx(func(_ userRepositoryTx, refreshRepo refreshTokenRepositoryTx) error {
		return s.persistRefreshToken(refreshRepo, userID, deviceID, replaceExistingDevice, refreshTokenModel)
	}); err != nil {
		return nil, nil, err
	}

	return tokens, refreshTokenModel, nil
}

func (s *authService) buildTokenPair(
	userID uint,
	role string,
	accessTokenVersion uint,
	deviceID, userAgent, ipAddress string,
	familyID string,
	rotatedFromTokenID *uint,
) (*AuthTokens, *tokenModule.RefreshToken, error) {
	expiry := time.Duration(s.Cfg.AccessTokenExpiryMinutes) * time.Minute
	if expiry == 0 {
		expiry = 15 * time.Minute // default fallback
	}

	accessToken, err := s.JWT.GenerateToken(userID, role, expiry, TokenAccess, accessTokenVersion)
	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := s.JWT.GenerateToken(userID, role, 7*24*time.Hour, TokenRefresh, accessTokenVersion)
	if err != nil {
		return nil, nil, err
	}
	if familyID == "" {
		familyID = uuid.NewString()
	}

	tokenHash := utils.HashToken(refreshToken)
	refreshTokenModel := &tokenModule.RefreshToken{
		UserID:             userID,
		TokenHash:          tokenHash,
		FamilyID:           familyID,
		RotatedFromTokenID: rotatedFromTokenID,
		DeviceID:           deviceID,
		UserAgent:          userAgent,
		IPAddress:          ipAddress,
		ExpiredAt:          time.Now().Add(7 * 24 * time.Hour),
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, refreshTokenModel, nil
}

func (s *authService) persistRefreshToken(
	refreshRepo refreshTokenRepositoryTx,
	userID uint,
	deviceID string,
	replaceExistingDevice bool,
	refreshTokenModel *tokenModule.RefreshToken,
) error {
	if replaceExistingDevice && deviceID != "" {
		if err := refreshRepo.DeleteByUserAndDevice(userID, deviceID); err != nil {
			return err
		}
	}
	return refreshRepo.Save(refreshTokenModel)
}
