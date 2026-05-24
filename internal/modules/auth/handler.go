package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"pleco-api/internal/httpx"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"pleco-api/internal/cache"
	"pleco-api/internal/erroroptimizer"
	"pleco-api/internal/modules/permission"
	"pleco-api/internal/modules/user"
	"pleco-api/internal/services"
)

type AuthHandler struct {
	AuthService    AuthService
	PermissionSvc  *permission.Service
	Cache          cache.Store
	ErrorOptimizer *erroroptimizer.ErrorOptimizerService
}

const (
	refreshTokenCookieName = "pleco_refresh_token"
	deviceIDCookieName     = "pleco_device_id"
)

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
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

	user := dtoToUser(input.Name, input.Email, input.PhoneNumber)
	err := h.AuthService.Register(&user, input.Password)
	if err != nil {
		if h.ErrorOptimizer != nil {
			language := c.GetHeader("Accept-Language")
			userContext := erroroptimizer.UserContext{
				Language:  language,
				IsNewUser: true,
			}

			optimized, errOpt := h.ErrorOptimizer.GetOptimizedError(
				c.Request.Context(),
				err,
				userContext,
				"/auth/register",
			)
			if errOpt == nil && optimized != nil {
				statusCode := http.StatusInternalServerError
				if optimized.Code == "AUTH_WEAK_PASSWORD" {
					statusCode = http.StatusBadRequest
				}
				c.JSON(statusCode, gin.H{
					"status":      "error",
					"code":        optimized.Code,
					"message":     optimized.Message,
					"details":     optimized.Details,
					"suggestions": optimized.Suggestions,
				})
				return
			}
		}

		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(strings.ToLower(err.Error()), "already in use") {
			httpx.ErrorWithCode(c, http.StatusBadRequest, httpx.ErrCodeEmailTaken, "Email already in use")
			return
		}
		if errors.Is(err, services.ErrWeakPassword) {
			httpx.ErrorWithCode(c, http.StatusBadRequest, httpx.ErrCodeWeakPassword, err.Error())
			return
		}
		httpx.ErrorWithCode(c, http.StatusInternalServerError, httpx.ErrCodeInternalError, "Failed to create user")
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

	deviceID := ensureDeviceID(c)
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()
	if input.DeviceName == "" {
		input.DeviceName = deviceID
	}

	tokens, err := h.AuthService.Login(input.Email, input.Password, deviceID, input.DeviceName, input.TrustedDevice, userAgent, ipAddress)
	if err != nil {
		if h.ErrorOptimizer != nil {
			language := c.GetHeader("Accept-Language")
			userContext := erroroptimizer.UserContext{
				Device:   deviceID,
				Language: language,
			}

			optimized, errOpt := h.ErrorOptimizer.GetOptimizedError(
				c.Request.Context(),
				err,
				userContext,
				"/auth/login",
			)
			if errOpt == nil && optimized != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"status":      "error",
					"code":        optimized.Code,
					"message":     optimized.Message,
					"details":     optimized.Details,
					"suggestions": optimized.Suggestions,
				})
				return
			}
		}

		if errors.Is(err, ErrInvalidCredentials) {
			httpx.ErrorWithCode(c, http.StatusUnauthorized, httpx.ErrCodeInvalidCredentials, "Invalid credentials")
			return
		}
		if errors.Is(err, ErrAccountLocked) {
			httpx.ErrorWithCode(c, http.StatusUnauthorized, httpx.ErrCodeAccountLocked, "Account locked due to too many failed attempts")
			return
		}
		if errors.Is(err, ErrEmailNotVerified) {
			httpx.ErrorWithCode(c, http.StatusUnauthorized, httpx.ErrCodeEmailNotVerified, "Please verify your email first")
			return
		}
		httpx.ErrorWithCode(c, http.StatusUnauthorized, httpx.ErrCodeInvalidCredentials, err.Error())
		return
	}

	setRefreshTokenCookie(c, tokens.RefreshToken)
	setDeviceIDCookie(c, tokens.DeviceID)
	httpx.Success(c, http.StatusOK, "Login success", accessTokenResponse{AccessToken: tokens.AccessToken}, nil)
}

func (h *AuthHandler) RequestOTP(c *gin.Context) {
	var input RequestOTPRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	err := h.AuthService.RequestOTP(c.Request.Context(), input.Channel, input.Target, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		status := http.StatusBadRequest
		message := "Unable to send OTP"
		if errors.Is(err, ErrOTPRateLimited) {
			status = http.StatusTooManyRequests
			message = "Too many OTP requests. Please try again later."
		}
		if errors.Is(err, ErrOTPWhatsAppTarget) {
			message = "No WhatsApp number is available for this account. Use email OTP or add a WhatsApp number in profile settings."
		}
		httpx.Error(c, status, message)
		return
	}

	httpx.Success(c, http.StatusOK, "OTP sent successfully", nil, nil)
}

func (h *AuthHandler) CheckPasswordlessIdentity(c *gin.Context) {
	var input PasswordlessIdentityRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}
	if err := h.AuthService.CheckPasswordlessIdentity(input.Channel, input.Target); err != nil {
		httpx.Error(c, http.StatusBadRequest, "Enter a valid email address or WhatsApp number.")
		return
	}
	httpx.Success(c, http.StatusOK, "Passwordless identity accepted", nil, nil)
}

func (h *AuthHandler) StartPasswordless(c *gin.Context) {
	var input PasswordlessIdentityRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	result, err := h.AuthService.StartPasswordless(c.Request.Context(), input.Channel, input.Target, ensureDeviceID(c), c.GetHeader("User-Agent"), c.ClientIP())
	if err != nil {
		status := http.StatusBadRequest
		message := "Unable to continue passwordless login"
		if errors.Is(err, ErrOTPRateLimited) {
			status = http.StatusTooManyRequests
			message = "Too many requests. Please try again later."
		}
		httpx.Error(c, status, message)
		return
	}

	httpx.Success(c, http.StatusOK, "Passwordless login started", result, nil)
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var input VerifyOTPRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	deviceID := ensureDeviceID(c)
	userAgent := c.GetHeader("User-Agent")
	if input.DeviceName == "" {
		input.DeviceName = deviceID
	}

	tokens, err := h.AuthService.VerifyOTP(c.Request.Context(), VerifyOTPInput{
		Channel:       input.Channel,
		Target:        input.Target,
		OTP:           input.OTP,
		DeviceID:      deviceID,
		DeviceName:    input.DeviceName,
		TrustedDevice: input.TrustedDevice,
		UserAgent:     userAgent,
		IPAddress:     c.ClientIP(),
	})
	if err != nil {
		httpx.ErrorWithCode(c, http.StatusUnauthorized, httpx.ErrCodeInvalidCredentials, "Invalid or expired OTP")
		return
	}

	setRefreshTokenCookie(c, tokens.RefreshToken)
	setDeviceIDCookie(c, tokens.DeviceID)
	httpx.Success(c, http.StatusOK, "OTP verified", accessTokenResponse{AccessToken: tokens.AccessToken}, nil)
}

func (h *AuthHandler) VerifyMagicLink(c *gin.Context) {
	var input VerifyMagicLinkRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	deviceID := ensureDeviceID(c)
	userAgent := c.GetHeader("User-Agent")
	if input.DeviceName == "" {
		input.DeviceName = deviceID
	}
	tokens, err := h.AuthService.VerifyMagicLink(input.Token, deviceID, input.DeviceName, input.TrustedDevice, userAgent, c.ClientIP())
	if err != nil {
		httpx.ErrorWithCode(c, http.StatusUnauthorized, httpx.ErrCodeInvalidCredentials, "Invalid or expired magic link")
		return
	}

	setRefreshTokenCookie(c, tokens.RefreshToken)
	setDeviceIDCookie(c, tokens.DeviceID)
	httpx.Success(c, http.StatusOK, "Magic link verified", accessTokenResponse{AccessToken: tokens.AccessToken}, nil)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	deviceID := currentDeviceID(c)
	if deviceID == "" {
		httpx.Error(c, http.StatusBadRequest, "device id required")
		return
	}

	err := h.AuthService.Logout(userID, deviceID)
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	clearRefreshTokenCookie(c)
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

	clearRefreshTokenCookie(c)
	httpx.Success(c, http.StatusOK, "all sessions revoked", nil, nil)
}

func (h *AuthHandler) LogoutOtherSessions(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	currentDeviceID := currentDeviceID(c)
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

	setRefreshTokenCookie(c, tokens.RefreshToken)
	setDeviceIDCookie(c, tokens.DeviceID)
	httpx.Success(c, http.StatusOK, "other sessions revoked", accessTokenResponse{AccessToken: tokens.AccessToken}, nil)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}

	if cookieRefreshToken, err := c.Cookie(refreshTokenCookieName); err == nil {
		body.RefreshToken = cookieRefreshToken
	} else if err := c.ShouldBindJSON(&body); err != nil {
		httpx.Error(c, http.StatusUnauthorized, "refresh token required")
		return
	}

	tokens, err := h.AuthService.RefreshToken(body.RefreshToken)
	if err != nil {
		clearRefreshTokenCookie(c)
		httpx.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	setRefreshTokenCookie(c, tokens.RefreshToken)
	setDeviceIDCookie(c, tokens.DeviceID)
	httpx.Success(c, http.StatusOK, "Refresh token success", accessTokenResponse{AccessToken: tokens.AccessToken}, nil)
}

func (h *AuthHandler) ListSessions(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	currentDeviceID := currentDeviceID(c)
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

func (h *AuthHandler) RevokeTrustedDevice(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	trustedDeviceID := c.Param("id")
	if trustedDeviceID == "" {
		httpx.Error(c, http.StatusBadRequest, "invalid trusted device id")
		return
	}

	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	if err := h.AuthService.RevokeTrustedDevice(userID, trustedDeviceID, userAgent, ipAddress); err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			httpx.Error(c, http.StatusNotFound, err.Error())
			return
		}
		httpx.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.Success(c, http.StatusOK, "trusted device removed", nil, nil)
}

func (h *AuthHandler) Profile(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if h.Cache != nil {
		var cached profileResponse
		key := fmt.Sprintf("user:profile:%d", userID)
		if ok, err := h.Cache.GetJSON(c.Request.Context(), key, &cached); err == nil && ok {
			httpx.Success(c, http.StatusOK, "Profile fetched", cached, nil)
			return
		}
	}

	user, err := h.AuthService.GetProfile(userID)
	if err != nil {
		httpx.Error(c, http.StatusNotFound, "User not found")
		return
	}

	permissions, _ := h.PermissionSvc.ListRolePermissionsByName(user.Role)

	response := profileResponse{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		PhoneNumber:   user.PhoneNumber,
		Role:          user.Role,
		IsVerified:    user.IsVerified,
		PhoneVerified: user.PhoneVerified,
		EmailVerified: user.EmailVerified,
		Permissions:   permissions,
	}
	if h.Cache != nil {
		_ = h.Cache.SetJSON(context.Background(), fmt.Sprintf("user:profile:%d", userID), response, 5*time.Minute)
	}

	httpx.Success(c, http.StatusOK, "Profile fetched", response, nil)
}

type profileResponse struct {
	ID            uint     `json:"id"`
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	PhoneNumber   string   `json:"phone_number,omitempty"`
	Role          string   `json:"role"`
	IsVerified    bool     `json:"is_verified"`
	PhoneVerified bool     `json:"phone_verified"`
	EmailVerified bool     `json:"email_verified"`
	Permissions   []string `json:"permissions"`
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
		httpx.ErrorWithCode(c, http.StatusBadRequest, httpx.ErrCodeInvalidCredentials, err.Error())
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
		httpx.ErrorWithCode(c, http.StatusInternalServerError, httpx.ErrCodeInternalError, err.Error())
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
		httpx.ErrorWithCode(c, http.StatusBadRequest, httpx.ErrCodeInvalidCredentials, err.Error())
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

	deviceID := ensureDeviceID(c)
	userAgent := c.GetHeader("User-Agent")
	ip := c.ClientIP()

	tokens, err := h.AuthService.SocialLogin(body.Provider, token, deviceID, userAgent, ip)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	setRefreshTokenCookie(c, tokens.RefreshToken)
	setDeviceIDCookie(c, tokens.DeviceID)
	httpx.Success(c, http.StatusOK, "Social login success", accessTokenResponse{AccessToken: tokens.AccessToken}, nil)
}

func (h *AuthHandler) SocialAccount(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	provider := c.Param("provider")
	cacheKey := fmt.Sprintf("social:account:%d:%s", userID, provider)
	if h.Cache != nil {
		var cached socialAccountResponse
		if ok, err := h.Cache.GetJSON(c.Request.Context(), cacheKey, &cached); err == nil && ok {
			httpx.Success(c, http.StatusOK, "Social account fetched", cached, nil)
			return
		}
	}

	account, err := h.AuthService.GetSocialAccount(userID, provider)
	if err != nil {
		httpx.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if account == nil {
		httpx.Error(c, http.StatusNotFound, "Social account not found")
		return
	}

	response := socialAccountResponse{
		ID:             account.ID,
		UserID:         account.UserID,
		Provider:       account.Provider,
		ProviderUserID: account.ProviderUserID,
		AvatarURL:      account.AvatarURL,
	}
	if h.Cache != nil {
		_ = h.Cache.SetJSON(context.Background(), cacheKey, response, 15*time.Minute)
	}

	httpx.Success(c, http.StatusOK, "Social account fetched", response, nil)
}

func dtoToUser(name, email, phoneNumber string) user.User {
	return user.User{
		Name:        name,
		Email:       email,
		PhoneNumber: phoneNumber,
		Role:        "user",
	}
}

type socialAccountResponse struct {
	ID             uint   `json:"id"`
	UserID         uint   `json:"user_id"`
	Provider       string `json:"provider"`
	ProviderUserID string `json:"provider_user_id"`
	AvatarURL      string `json:"avatar_url"`
}

func setRefreshTokenCookie(c *gin.Context, refreshToken string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
}

func ensureDeviceID(c *gin.Context) string {
	if deviceID := currentDeviceID(c); deviceID != "" {
		setDeviceIDCookie(c, deviceID)
		return deviceID
	}
	deviceID := "device-" + uuid.NewString()
	setDeviceIDCookie(c, deviceID)
	return deviceID
}

func currentDeviceID(c *gin.Context) string {
	if deviceID, err := c.Cookie(deviceIDCookieName); err == nil && deviceID != "" {
		return deviceID
	}
	return c.GetHeader("X-Device-ID")
}

func setDeviceIDCookie(c *gin.Context, deviceID string) {
	if deviceID == "" {
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     deviceIDCookieName,
		Value:    deviceID,
		Path:     "/",
		MaxAge:   int((365 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
}

func clearRefreshTokenCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
}
