package utils

import (
	"fmt"
	user "go-api-starterkit/modules/user"

	"github.com/gin-gonic/gin"
)

// getUserIDFromContext safely extracts the user_id as uint from the Gin context.
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	val, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	switch v := val.(type) {
	case uint:
		return v, true
	case int:
		return uint(v), true
	case float64:
		return uint(v), true
	default:
		return 0, false
	}
}

// dtoToUser converts registration fields to user.User (excluding hashed password).
func DtoToUser(name, email string) user.User {
	return user.User{
		Name:  name,
		Email: email,
		Role:  "user",
	}
}

// GetUserIDFromToken extracts the user_id as uint from a JWT claims map.
func GetUserIDFromToken(claims map[string]interface{}) (uint, bool) {
	val, exists := claims["user_id"]
	if !exists {
		return 0, false
	}
	switch v := val.(type) {
	case uint:
		return v, true
	case int:
		return uint(v), true
	case float64:
		return uint(v), true
	case string:
		// Attempt to parse uint from string
		var id uint
		_, err := fmt.Sscanf(v, "%d", &id)
		if err == nil {
			return id, true
		}
		return 0, false
	default:
		return 0, false
	}
}
