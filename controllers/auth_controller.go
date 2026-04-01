package controllers

import (
	"net/http"

	"go-auth-app/dto"
	"go-auth-app/services"
	"go-auth-app/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	AuthService services.AuthService
}

func (a *AuthController) Register(c *gin.Context) {
	var input dto.RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Delegate password hashing and uniqueness check to the AuthService.
	user := utils.DtoToUser(&input)
	err := a.AuthService.Register(&user, input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered"})
}

func (a *AuthController) Login(c *gin.Context) {
	var input dto.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	deviceID := c.GetHeader("X-Device-ID")
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	tokens, err := a.AuthService.Login(input.Email, input.Password, deviceID, userAgent, ipAddress)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (a *AuthController) Logout(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device id required"})
		return
	}

	err := a.AuthService.Logout(userID, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logout success"})
}

func (a *AuthController) RefreshToken(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	tokens, err := a.AuthService.RefreshToken(body.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (a *AuthController) Profile(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := a.AuthService.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	})
}

func (h *AuthController) VerifyEmail(c *gin.Context) {
	token := c.Query("token")

	if token == "" {
		c.JSON(400, gin.H{"error": "token required"})
		return
	}

	err := h.AuthService.VerifyEmail(token)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "email verified"})
}

func (a *AuthController) ResendVerification(c *gin.Context) {
	type ResendInput struct {
		Email string `json:"email" binding:"required,email"`
	}

	var input ResendInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email"})
		return
	}

	err := a.AuthService.ResendVerification(input.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification email resent"})
}
