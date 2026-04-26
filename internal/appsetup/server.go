package appsetup

import (
	"pleco-api/internal/config"
	"pleco-api/internal/services"

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
	router, err := BuildRouter(db, appConfig, jwtService)
	if err != nil {
		return err
	}

	if registerDocs != nil {
		registerDocs(router)
	}

	return router.Run(":" + appConfig.Port)
}
