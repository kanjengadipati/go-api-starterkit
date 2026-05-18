package domain

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmailNotVerified    = errors.New("please verify your email first")
	ErrSessionNotFound     = errors.New("session not found")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidTokenType    = errors.New("invalid token type")
	ErrInvalidTokenClaims  = errors.New("invalid token")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrRefreshTokenReuse   = errors.New("refresh token reuse detected")
	ErrAccountLocked       = errors.New("account locked due to too many failed attempts")
)
