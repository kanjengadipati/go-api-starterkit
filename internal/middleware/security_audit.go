package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type SecurityAuditEvent struct {
	UserID      *uint
	Method      string
	Path        string
	Status      int
	IPAddress   string
	UserAgent   string
	Description string
}

func SecurityAuditLogger(record func(SecurityAuditEvent)) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if record == nil || !isSecurityRelevantStatus(c.Writer.Status()) || !strings.HasPrefix(c.Request.URL.Path, "/auth") {
			return
		}

		var userID *uint
		if value, ok := c.Get("user_id"); ok {
			if id, ok := value.(uint); ok {
				userID = &id
			}
		}

		record(SecurityAuditEvent{
			UserID:      userID,
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			Status:      c.Writer.Status(),
			Description: c.Request.Method + " " + c.Request.URL.Path,
			IPAddress:   c.ClientIP(),
			UserAgent:   c.GetHeader("User-Agent"),
		})
	}
}

func isSecurityRelevantStatus(status int) bool {
	return status == http.StatusUnauthorized ||
		status == http.StatusForbidden ||
		status == http.StatusTooManyRequests
}
