package role

import (
	"net/http"

	"go-api-starterkit/httpx"

	"github.com/gin-gonic/gin"
)

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists || userRole != role {
			httpx.Error(c, http.StatusForbidden, "Forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}
