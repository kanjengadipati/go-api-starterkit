package middleware

import (
	"net/http"

	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
)

// AccessTokenVersionSource returns the current access-token security version for a user.
// Implemented in appsetup (e.g. wrapping user.Repository) to avoid import cycles.
type AccessTokenVersionSource interface {
	AccessTokenVersionForUser(userID uint) (uint, error)
}

// RequireAccessTokenVersion rejects access tokens whose embedded `tv` claim no longer
// matches the database (e.g. after a role change or password reset). Use after AuthMiddleware.
func RequireAccessTokenVersion(src AccessTokenVersionSource) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := c.Get("user_id")
		if !ok {
			httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}
		userID, ok := uid.(uint)
		if !ok {
			httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		claimVer, ok := accessTokenVersionFromContext(c)
		if !ok {
			httpx.Error(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		dbVer, err := src.AccessTokenVersionForUser(userID)
		if err != nil {
			httpx.Error(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		if dbVer != claimVer {
			httpx.Error(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		c.Next()
	}
}

func accessTokenVersionFromContext(c *gin.Context) (uint, bool) {
	raw, exists := c.Get("access_token_version")
	if !exists {
		return 0, true
	}
	switch v := raw.(type) {
	case uint:
		return v, true
	case float64:
		return uint(v), true
	case int:
		if v < 0 {
			return 0, false
		}
		return uint(v), true
	default:
		return 0, false
	}
}
