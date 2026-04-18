package auth

import (
	"go-api-starterkit/middleware"
	"go-api-starterkit/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(api *gin.RouterGroup, handler *AuthHandler, jwtService *services.JWTService) {
	auth := api.Group("/auth")
	auth.POST("/register", handler.Register)
	auth.POST("/login", handler.Login)
	auth.POST("/refresh", handler.RefreshToken)
	auth.GET("/verify", handler.VerifyEmail)
	auth.GET("/resend-verification", handler.ResendVerification)
	auth.POST("/forgot-password", handler.ForgotPassword)
	auth.POST("/reset-password", handler.ResetPassword)
	auth.POST("/social-login", handler.SocialLogin)

	protected := auth.Group("/")
	protected.Use(middleware.AuthMiddleware(jwtService))
	protected.GET("/profile", handler.Profile)
	protected.POST("/logout", handler.Logout)
}
