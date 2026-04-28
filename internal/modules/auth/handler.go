package auth

import (
	"errors"
	"net/http"
	"pleco-api/internal/httpx"
	"strconv"

	"github.com/gin-gonic/gin"
	"pleco-api/internal/modules/permission"
	"pleco-api/internal/modules/user"
	"pleco-api/internal/services"
)

type AuthHandler struct {
	AuthService   AuthService
	PermissionSvc *permission.Service
}

func NewHandler(authService AuthService, permissionSvc *permission.Service) *AuthHandler {
	return &AuthHandler{AuthService: authService, PermissionSvc: permissionSvc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	user := dtoToUser(input.Name, input.Email)
	err := h.AuthService.Register(&user, input.Password)
	if err != nil {
		if errors.Is(err, services.ErrWeakPassword) {
			httpx.Error(c, http.StatusBadRequest, err.Error())
			return
		}
		httpx.Error(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	httpx.Success(c, http.StatusOK, "User registered", nil, nil)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	deviceID := c.GetHeader("X-Device-ID")
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	tokens, err := h.AuthService.Login(input.Email, input.Password, deviceID, userAgent, ipAddress)
	if err != nil {
		httpx.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	httpx.Success(c, http.StatusOK, "Login success", tokens, nil)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	deviceID := c.GetHeader("X-Device-ID")
	if deviceID == "" {
		httpx.Error(c, http.StatusBadRequest, "device id required")
		return
	}

	err := h.AuthService.Logout(userID, deviceID)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.Success(c, http.StatusOK, "logout success", nil, nil)
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	if err := h.AuthService.LogoutAll(userID, userAgent, ipAddress); err != nil {
		httpx.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.Success(c, http.StatusOK, "all sessions revoked", nil, nil)
}

func (h *AuthHandler) LogoutOtherSessions(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	currentDeviceID := c.GetHeader("X-Device-ID")
	if currentDeviceID == "" {
		httpx.Error(c, http.StatusBadRequest, "device id required")
		return
	}

	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	tokens, err := h.AuthService.LogoutOtherSessions(userID, currentDeviceID, userAgent, ipAddress)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.Success(c, http.StatusOK, "other sessions revoked", tokens, nil)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	tokens, err := h.AuthService.RefreshToken(body.RefreshToken)
	if err != nil {
		httpx.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	httpx.Success(c, http.StatusOK, "Refresh token success", tokens, nil)
}

func (h *AuthHandler) ListSessions(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	currentDeviceID := c.GetHeader("X-Device-ID")
	sessions, err := h.AuthService.ListSessions(userID, currentDeviceID)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, "Failed to fetch sessions")
		return
	}

	httpx.Success(c, http.StatusOK, "sessions fetched", sessions, nil)
}

func (h *AuthHandler) RevokeSession(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, "invalid session id")
		return
	}

	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	if err := h.AuthService.RevokeSession(userID, uint(sessionID), userAgent, ipAddress); err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			httpx.Error(c, http.StatusNotFound, err.Error())
			return
		}
		httpx.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.Success(c, http.StatusOK, "session revoked", nil, nil)
}

func (h *AuthHandler) Profile(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.AuthService.GetProfile(userID)
	if err != nil {
		httpx.Error(c, http.StatusNotFound, "User not found")
		return
	}

	permissions, _ := h.PermissionSvc.Repo.ListRolePermissionsByName(user.Role)

	httpx.Success(c, http.StatusOK, "Profile fetched", gin.H{
		"id":          user.ID,
		"name":        user.Name,
		"email":       user.Email,
		"role":        user.Role,
		"permissions": permissions,
	}, nil)
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")

	if token == "" {
		httpx.Error(c, http.StatusBadRequest, "token required")
		return
	}

	err := h.AuthService.VerifyEmail(token)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	httpx.Success(c, http.StatusOK, "email verified", nil, nil)
}

func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var input ResendVerificationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	err := h.AuthService.ResendVerification(input.Email)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	httpx.Success(c, http.StatusOK, "Verification email resent", nil, nil)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var body ForgotPasswordRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	err := h.AuthService.ForgotPassword(body.Email)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	httpx.Success(c, http.StatusOK, "reset link sent", nil, nil)
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var body ResetPasswordRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	err := h.AuthService.ResetPassword(body.Token, body.NewPassword)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	httpx.Success(c, http.StatusOK, "password updated", nil, nil)
}

func (h *AuthHandler) SocialLogin(c *gin.Context) {
	var body SocialLoginRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	token := body.Token
	if token == "" {
		token = body.IDToken
	}
	if token == "" {
		httpx.Error(c, http.StatusBadRequest, "social token required")
		return
	}

	deviceID := c.GetHeader("X-Device-ID")
	userAgent := c.GetHeader("User-Agent")
	ip := c.ClientIP()

	tokens, err := h.AuthService.SocialLogin(body.Provider, token, deviceID, userAgent, ip)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	httpx.Success(c, http.StatusOK, "Social login success", tokens, nil)
}

func dtoToUser(name, email string) user.User {
	return user.User{
		Name:  name,
		Email: email,
		Role:  "user",
	}
}
