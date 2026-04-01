package routes

import (
	"go-auth-app/config"
	"go-auth-app/controllers"
	"go-auth-app/middleware"
	"go-auth-app/repositories"
	"go-auth-app/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	userRepo := &repositories.UserRepoDB{}
	refreshTokenRepo := repositories.NewRefreshTokenRepo()
	emailVerificationRepo := repositories.NewEmailVerificationTokenRepo()
	jwtService := services.NewJWTService(config.JWTSecret)
	emailSvc := services.NewEmailService()

	authService := services.NewAuthService(
		userRepo,
		refreshTokenRepo,
		emailVerificationRepo,
		jwtService,
		emailSvc,
	)

	userService := &services.UserService{
		UserRepo: userRepo,
	}

	authController := controllers.AuthController{
		AuthService: authService,
	}

	userController := controllers.UserController{
		UserService: userService,
	}

	api := router.Group("/api")

	// ========================
	// 🔓 PUBLIC ROUTES
	// ========================
	api.POST("/register", authController.Register)
	api.POST("/login", authController.Login)
	api.POST("/refresh", authController.RefreshToken)
	api.GET("/verify", authController.VerifyEmail)
	api.GET("/resend-verification", authController.ResendVerification)

	// ========================
	// 🔐 PROTECTED ROUTES
	// ========================
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(jwtService))

	protected.GET("/profile", authController.Profile)
	protected.POST("/logout", authController.Logout)

	// ========================
	// 👑 ADMIN ROUTES
	// ========================
	admin := protected.Group("/admin")
	admin.Use(middleware.AdminOnly())

	admin.GET("/users", userController.GetAllUsers)
	admin.DELETE("/users/:id", userController.DeleteUser)
}
