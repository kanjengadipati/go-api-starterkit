package main

import (
	"go-auth-app/config"
	"go-auth-app/modules/auth"
	"go-auth-app/modules/user"

	"github.com/gin-gonic/gin"
)

func initApp() *gin.Engine {
	// Load env & init JWT
	config.LoadEnv()
	config.InitJWT()

	// Connect DB
	db := config.ConnectDB()
	_ = db

	router := gin.Default()
	api := router.Group("/")

	// ===== USER =====
	userRepo := user.NewRepository()
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// ===== AUTH =====
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, userService)
	authHandler := auth.NewHandler(authService)

	// ===== ROUTES =====
	auth.SetupRoutes(api, authHandler)
	user.SetupRoutes(api, userHandler)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return router
}

func main() {
	router := initApp()

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
