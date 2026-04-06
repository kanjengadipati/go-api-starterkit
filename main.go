package main

import (
	"go-auth-app/config"
	"go-auth-app/routes"

	"github.com/gin-gonic/gin"
)

func initApp() *gin.Engine {
	// Load environment variables and initialize JWT
	config.LoadEnv()
	config.InitJWT()

	// Connect to the database
	config.ConnectDB()

	// Initialize Gin router and set up routes
	router := gin.Default()
	routes.SetupRoutes(router)

	return router
}

func main() {
	router := initApp()
	// Start the server
	if err := router.Run(":8080"); err != nil {
		panic("failed to start server: " + err.Error())
	}
}
