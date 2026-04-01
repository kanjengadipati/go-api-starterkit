package services

import (
	"errors"
	"log"

	"go-auth-app/models"
	"go-auth-app/repositories"
	"go-auth-app/utils"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(user *models.User, password string) error
	Login(email, password, deviceID, userAgent, ipAddress string) (*AuthTokens, error)
	Logout(userID uint, deviceID string) error
	RefreshToken(oldRefreshToken string) (*AuthTokens, error)
	GetProfile(userID uint) (*models.User, error)
	ResendVerification(email string) error
	VerifyEmail(token string) error
}

type authService struct {
	UserRepo              repositories.UserRepository
	RefreshTokenRepo      repositories.RefreshTokenRepository
	EmailVerificationRepo repositories.EmailVerificationTokenRepository
	JWT                   *JWTService
	EmailSvc              EmailService
}

var _ AuthService = (*authService)(nil)

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

const (
	TokenAccess  = "access"
	TokenRefresh = "refresh"
)

func NewAuthService(
	userRepo repositories.UserRepository,
	refreshRepo repositories.RefreshTokenRepository,
	emailRepo repositories.EmailVerificationTokenRepository,
	jwt *JWTService,
	emailSvc EmailService,
) AuthService {
	return &authService{
		UserRepo:              userRepo,
		RefreshTokenRepo:      refreshRepo,
		EmailVerificationRepo: emailRepo,
		JWT:                   jwt,
		EmailSvc:              emailSvc,
	}
}

func (s *authService) Register(user *models.User, password string) error {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	user.Role = "user"
	user.IsVerified = false

	// Save the user to the repository first (so user.ID is set for tokens)
	if err := s.UserRepo.Create(user); err != nil {
		return err
	}

	// Generate email verification token
	verificationToken := uuid.NewString()
	verificationRecord := &models.EmailVerificationToken{
		UserID:    user.ID,
		Token:     verificationToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	// Save email verification token to DB for validation

	if err := s.EmailVerificationRepo.Create(verificationRecord); err != nil {
		log.Printf("failed to save email verification token for %s: %v", user.Email, err)
	}

	// Send verification email; log error if sending fails but do not delete user
	if sendErr := s.EmailSvc.SendVerificationEmail(user.Email, verificationToken); sendErr != nil {
		log.Printf("Failed to send verification email to %s: %v", user.Email, sendErr)
		return sendErr
	}

	return nil
}

func (s *authService) Login(email, password, deviceID, userAgent, ipAddress string) (*AuthTokens, error) {
	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsVerified {
		return nil, errors.New("please verify your email first")
	}

	accessToken, err := s.JWT.GenerateToken(user.ID, user.Role, 15*time.Minute, TokenAccess)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.JWT.GenerateToken(user.ID, user.Role, 7*24*time.Hour, TokenRefresh)
	if err != nil {
		return nil, err
	}

	tokenHash := utils.HashToken(refreshToken)

	refreshTokenModel := &models.RefreshToken{
		UserID:    user.ID,
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

func (s *authService) Logout(userID uint, deviceID string) error {

	token, err := s.RefreshTokenRepo.FindByUserAndDevice(userID, deviceID)
	if err != nil {
		return err
	}

	return s.RefreshTokenRepo.DeleteByID(token.ID)
}

func (s *authService) LogoutAll(userID uint) error {
	return s.RefreshTokenRepo.DeleteByUser(userID)
}

func (s *authService) RefreshToken(oldRefreshToken string) (*AuthTokens, error) {
	claims, err := s.JWT.ValidateToken(oldRefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check token type
	if claims["type"] != TokenRefresh {
		return nil, errors.New("invalid token type")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("invalid user id in token claims")
	}
	uid := uint(userID)

	// Hash old refresh token and find matching token in DB
	tokens, err := s.RefreshTokenRepo.FindByUser(uid)
	if err != nil {
		return nil, err
	}
	oldHash := utils.HashToken(oldRefreshToken)

	var matchedToken *models.RefreshToken
	for i := range tokens {
		if tokens[i].TokenHash == oldHash {
			matchedToken = &tokens[i]
			break
		}
	}
	if matchedToken == nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check expiration
	if time.Now().After(matchedToken.ExpiredAt) {
		return nil, errors.New("refresh token expired")
	}

	// Rotation: remove old token
	if err := s.RefreshTokenRepo.DeleteByID(matchedToken.ID); err != nil {
		return nil, err
	}

	// Generate new tokens
	user, err := s.UserRepo.FindByID(uid)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.JWT.GenerateToken(uid, user.Role, 15*time.Minute, TokenAccess)
	if err != nil {
		return nil, err
	}
	newRefreshToken, err := s.JWT.GenerateToken(uid, user.Role, 7*24*time.Hour, TokenRefresh)
	if err != nil {
		return nil, err
	}

	newHash := utils.HashToken(newRefreshToken)
	newToken := &models.RefreshToken{
		UserID:    uid,
		TokenHash: newHash,
		DeviceID:  matchedToken.DeviceID,
		UserAgent: matchedToken.UserAgent,
		IPAddress: matchedToken.IPAddress,
		ExpiredAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.RefreshTokenRepo.Save(newToken); err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *authService) GetProfile(userID uint) (*models.User, error) {
	return s.UserRepo.FindByID(userID)
}

func (s *authService) ResendVerification(email string) error {
	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		return err
	}

	if user.IsVerified {
		return errors.New("already verified")
	}

	token := uuid.NewString()

	_ = s.EmailVerificationRepo.DeleteByUserID(user.ID)

	verification := &models.EmailVerificationToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	_ = s.EmailVerificationRepo.Create(verification)

	return s.EmailSvc.SendVerificationEmail(user.Email, token)
}

func (s *authService) VerifyEmail(token string) error {
	verification, err := s.EmailVerificationRepo.FindByToken(token)
	if err != nil {
		return errors.New("invalid token")
	}

	if time.Now().After(verification.ExpiresAt) {
		return errors.New("token expired")
	}

	user, err := s.UserRepo.FindByID(verification.UserID)
	if err != nil {
		return err
	}

	user.IsVerified = true
	if err := s.UserRepo.Update(user); err != nil {
		return err
	}

	// delete token
	_ = s.EmailVerificationRepo.DeleteByID(verification.ID)

	return nil
}
