package auth

import (
	"go-api-starterkit/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	AuthService AuthService
}

func NewHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, utils.FormatValidationError(err))
		return
	}

	user := utils.DtoToUser(input.Name, input.Email)
	err := h.AuthService.Register(&user, input.Password)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	utils.Success(c, http.StatusOK, "User registered", nil, nil)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, utils.FormatValidationError(err))
		return
	}

	deviceID := c.GetHeader("X-Device-ID")
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	tokens, err := h.AuthService.Login(input.Email, input.Password, deviceID, userAgent, ipAddress)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Login success", tokens, nil)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		utils.Error(c, http.StatusBadRequest, "device id required")
		return
	}

	err := h.AuthService.Logout(userID, deviceID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "logout success", nil, nil)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ValidationError(c, utils.FormatValidationError(err))
		return
	}

	tokens, err := h.AuthService.RefreshToken(body.RefreshToken)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Refresh token success", tokens, nil)
}

func (h *AuthHandler) Profile(c *gin.Context) {
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.AuthService.GetProfile(userID)
	if err != nil {
		utils.Error(c, http.StatusNotFound, "User not found")
		return
	}

	utils.Success(c, http.StatusOK, "Profile fetched", gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	}, nil)
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")

	if token == "" {
		utils.Error(c, http.StatusBadRequest, "token required")
		return
	}

	err := h.AuthService.VerifyEmail(token)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "email verified", nil, nil)
}

func (h *AuthHandler) ResendVerification(c *gin.Context) {
	type ResendInput struct {
		Email string `json:"email" binding:"required,email"`
	}

	var input ResendInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, utils.FormatValidationError(err))
		return
	}

	err := h.AuthService.ResendVerification(input.Email)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Verification email resent", nil, nil)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var body struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ValidationError(c, utils.FormatValidationError(err))
		return
	}

	err := h.AuthService.ForgotPassword(body.Email)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "reset link sent", nil, nil)
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var body struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ValidationError(c, utils.FormatValidationError(err))
		return
	}

	err := h.AuthService.ResetPassword(body.Token, body.NewPassword)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "password updated", nil, nil)
}

func (h *AuthHandler) SocialLogin(c *gin.Context) {
	var body struct {
		Provider string `json:"provider"`
		Token    string `json:"id_token"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ValidationError(c, utils.FormatValidationError(err))
		return
	}

	deviceID := "web"
	userAgent := c.GetHeader("User-Agent")
	ip := c.ClientIP()

	tokens, err := h.AuthService.SocialLogin(body.Provider, body.Token, deviceID, userAgent, ip)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, http.StatusOK, "Social login success", tokens, nil)
}
