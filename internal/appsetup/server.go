package appsetup

import (
	"go-api-starterkit/internal/config"
	"go-api-starterkit/internal/services"

	"github.com/gin-gonic/gin"
)

func RunAPI(registerDocs func(*gin.Engine)) error {
	config.LoadEnv()

	appConfig := config.LoadAppConfig()
	if err := appConfig.Validate(); err != nil {
		return err
	}

	db := config.ConnectDB(appConfig.DatabaseURL)
	RunStartupTasks(appConfig, db)

	jwtService := services.NewJWTService(appConfig.JWTSecret)
	router := BuildRouter(db, appConfig, jwtService)

	if registerDocs != nil {
		registerDocs(router)
	}

	return router.Run(":" + appConfig.Port)
}
