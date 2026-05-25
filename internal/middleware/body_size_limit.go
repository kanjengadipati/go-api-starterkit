package middleware

import (
	"net/http"

	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
)

func BodySizeLimit(maxBytes int64) gin.HandlerFunc {
	if maxBytes < 1 {
		maxBytes = 1 << 20
	}
	return func(c *gin.Context) {
		if c.Request != nil && c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
		for _, err := range c.Errors {
			if err != nil && err.Err != nil && err.Err.Error() == "http: request body too large" {
				httpx.ErrorWithCode(c, http.StatusRequestEntityTooLarge, "REQUEST_BODY_TOO_LARGE", "Request body is too large")
				c.Abort()
				return
			}
		}
	}
}
