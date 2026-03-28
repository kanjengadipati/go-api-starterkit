package utils

import (
	"go-auth-app/dto"
	"go-auth-app/models"

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

// dtoToUser converts dto.RegisterRequest to models.User (excluding hashed password).
func DtoToUser(input *dto.RegisterRequest) models.User {
	return models.User{
		Name:  input.Name,
		Email: input.Email,
		Role:  "user",
	}
}
