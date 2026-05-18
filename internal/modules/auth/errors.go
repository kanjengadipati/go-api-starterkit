package auth

import "pleco-api/internal/domain"

var (
	ErrInvalidCredentials  = domain.ErrInvalidCredentials
	ErrEmailNotVerified    = domain.ErrEmailNotVerified
	ErrSessionNotFound     = domain.ErrSessionNotFound
	ErrInvalidRefreshToken = domain.ErrInvalidRefreshToken
	ErrInvalidTokenType    = domain.ErrInvalidTokenType
	ErrInvalidTokenClaims  = domain.ErrInvalidTokenClaims
	ErrRefreshTokenExpired = domain.ErrRefreshTokenExpired
	ErrRefreshTokenReuse   = domain.ErrRefreshTokenReuse
	ErrAccountLocked       = domain.ErrAccountLocked
)
