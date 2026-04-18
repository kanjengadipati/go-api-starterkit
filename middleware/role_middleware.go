package middleware

import (
	"net/http"

	"go-api-starterkit/httpx"

	"github.com/gin-gonic/gin"
)

type permissionChecker interface {
	HasPermission(roleName, permission string) (bool, error)
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")

		if !exists || role != "admin" {
			httpx.Error(c, http.StatusForbidden, "Forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}

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

func RequirePermission(checker permissionChecker, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		roleName, ok := roleValue.(string)
		if !exists || !ok || roleName == "" {
			httpx.Error(c, http.StatusForbidden, "Forbidden")
			c.Abort()
			return
		}

		allowed, err := checker.HasPermission(roleName, permission)
		if err != nil {
			httpx.Error(c, http.StatusInternalServerError, "Failed to check permissions")
			c.Abort()
			return
		}

		if !allowed {
			httpx.Error(c, http.StatusForbidden, "Forbidden")
			c.Abort()
			return
		}

		c.Next()
	}
}
