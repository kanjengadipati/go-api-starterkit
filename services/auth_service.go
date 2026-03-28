package services

import (
	"errors"

	"go-auth-app/models"
	"go-auth-app/repositories"
	"go-auth-app/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepo         repositories.UserRepository
	RefreshTokenRepo repositories.RefreshTokenRepository
	JWT              *JWTService
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

const (
	TokenAccess  = "access"
	TokenRefresh = "refresh"
)

func (s *AuthService) Register(user *models.User, password string) error {
	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	user.Role = "user"
	return s.UserRepo.Create(user)
}

func (s *AuthService) Login(email, password, deviceID, userAgent, ipAddress string) (*AuthTokens, error) {
	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
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

func (s *AuthService) Logout(userID uint, deviceID string) error {

	token, err := s.RefreshTokenRepo.FindByUserAndDevice(userID, deviceID)
	if err != nil {
		return err
	}

	return s.RefreshTokenRepo.DeleteByID(token.ID)
}

func (s *AuthService) LogoutAll(userID uint) error {
	return s.RefreshTokenRepo.DeleteByUser(userID)
}

func (s *AuthService) RefreshToken(oldRefreshToken string) (*AuthTokens, error) {
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

func (s *AuthService) GetProfile(userID uint) (*models.User, error) {
	user, err := s.UserRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// contoh future logic
	// - audit log
	// - enrich data
	// - cache

	return user, nil
}
