package services

import (
	"errors"

	"go-auth-app/models"
	"go-auth-app/repositories"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserRepo repositories.UserRepository
	JWT      *JWTService
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

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

func (s *AuthService) Login(email, password string) (*AuthTokens, error) {
	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	accessToken, err := s.JWT.GenerateToken(user.ID, user.Role, time.Minute*15)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.JWT.GenerateToken(user.ID, user.Role, time.Hour*24*7)
	if err != nil {
		return nil, err
	}

	user.RefreshToken = refreshToken
	err = s.UserRepo.Update(user)
	if err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Logout(userID uint, refreshToken string) error {
	user, err := s.UserRepo.FindByID(userID)
	if err != nil {
		return err
	}

	// optional: validasi refresh token
	if user.RefreshToken != refreshToken {
		return errors.New("invalid refresh token")
	}

	// hapus refresh token (invalidate session)
	user.RefreshToken = ""

	return s.UserRepo.Update(user)
}

func (s *AuthService) RefreshToken(oldRefreshToken string) (*AuthTokens, error) {
	// ✅ validate token via JWT service
	claims, err := s.JWT.ValidateToken(oldRefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("invalid token")
	}

	userID := uint(userIDFloat)

	// ✅ ambil user dari repo
	user, err := s.UserRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// ✅ validasi refresh token match DB
	if user.RefreshToken != oldRefreshToken {
		return nil, errors.New("invalid refresh token")
	}

	// ✅ generate new tokens
	accessToken, err := s.JWT.GenerateToken(user.ID, user.Role, time.Minute*15)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.JWT.GenerateToken(user.ID, user.Role, time.Hour*24*7)
	if err != nil {
		return nil, err
	}

	// ✅ rotate refresh token
	user.RefreshToken = refreshToken
	err = s.UserRepo.Update(user)
	if err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
