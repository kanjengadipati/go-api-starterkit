package auth

import (
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *AuthHandler, jwtService *services.JWTService, rateStore middleware.RateLimitStore) {
	auth := api.Group("/auth")
	if rateStore == nil {
		rateStore = middleware.NewInMemoryRateLimitStore()
	}
	loginLimiter := middleware.NewRateLimiterWithStore(5, time.Minute, rateStore)
	registerLimiter := middleware.NewRateLimiterWithStore(5, time.Minute, rateStore)
	passwordLimiter := middleware.NewRateLimiterWithStore(3, 5*time.Minute, rateStore)
	refreshLimiter := middleware.NewRateLimiterWithStore(10, time.Minute, rateStore)
	socialLimiter := middleware.NewRateLimiterWithStore(5, time.Minute, rateStore)

	auth.POST("/register", registerLimiter.Middleware(), handler.Register)
	auth.POST("/login", loginLimiter.Middleware(), handler.Login)
	auth.POST("/refresh", refreshLimiter.Middleware(), handler.RefreshToken)
	auth.GET("/verify", handler.VerifyEmail)
	auth.POST("/resend-verification", passwordLimiter.Middleware(), handler.ResendVerification)
	auth.POST("/forgot-password", passwordLimiter.Middleware(), handler.ForgotPassword)
	auth.POST("/reset-password", passwordLimiter.Middleware(), handler.ResetPassword)
	auth.POST("/social-login", socialLimiter.Middleware(), handler.SocialLogin)

	protected := auth.Group("/")
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.GET("/profile", handler.Profile)
	protected.GET("/sessions", handler.ListSessions)
	protected.POST("/logout", handler.Logout)
	protected.POST("/logout-all", handler.LogoutAll)
	protected.POST("/logout-others", handler.LogoutOtherSessions)
	protected.DELETE("/sessions/:id", handler.RevokeSession)
}
