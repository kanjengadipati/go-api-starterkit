package domain

import "errors"

type ErrorCode string

const (
	CodeInvalidRequest    ErrorCode = "INVALID_REQUEST"
	CodeValidationFailed  ErrorCode = "VALIDATION_FAILED"
	CodeUnauthenticated   ErrorCode = "UNAUTHENTICATED"
	CodeForbidden         ErrorCode = "FORBIDDEN"
	CodeNotFound          ErrorCode = "NOT_FOUND"
	CodeRateLimited       ErrorCode = "RATE_LIMIT_EXCEEDED"
	CodeConflict          ErrorCode = "CONFLICT"
	CodeInternal          ErrorCode = "SERVER_INTERNAL_ERROR"
	CodeInvalidCredential ErrorCode = "AUTH_INVALID_CREDENTIALS"
	CodeWeakPassword      ErrorCode = "AUTH_WEAK_PASSWORD"
)

type APIError struct {
	Code       ErrorCode
	Message    string
	HTTPStatus int
	Err        error
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *APIError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewAPIError(status int, code ErrorCode, message string, err error) *APIError {
	return &APIError{HTTPStatus: status, Code: code, Message: message, Err: err}
}

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
