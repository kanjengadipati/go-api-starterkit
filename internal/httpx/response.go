package httpx

import "github.com/gin-gonic/gin"

const (
	ErrCodeInvalidCredentials = "AUTH_INVALID_CREDENTIALS"
	ErrCodeEmailTaken         = "AUTH_EMAIL_TAKEN"
	ErrCodeWeakPassword       = "AUTH_WEAK_PASSWORD"
	ErrCodeAccountLocked      = "AUTH_ACCOUNT_LOCKED"
	ErrCodeEmailNotVerified   = "AUTH_EMAIL_NOT_VERIFIED"
	ErrCodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	ErrCodeInternalError      = "SERVER_INTERNAL_ERROR"
)

type Envelope struct {
	Status  string      `json:"status"`
	Code    string      `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

func Respond(c *gin.Context, code int, status, message string, data, meta, errors interface{}) {
	c.JSON(code, Envelope{
		Status:  status,
		Message: message,
		Data:    data,
		Meta:    meta,
		Errors:  errors,
	})
}

func Success(c *gin.Context, code int, message string, data interface{}, meta interface{}) {
	Respond(c, code, "success", message, data, meta, nil)
}

func Error(c *gin.Context, code int, message string) {
	Respond(c, code, "error", message, nil, nil, nil)
}

func ErrorWithCode(c *gin.Context, code int, errorCode string, message string) {
	c.JSON(code, Envelope{
		Status:  "error",
		Code:    errorCode,
		Message: message,
	})
}
