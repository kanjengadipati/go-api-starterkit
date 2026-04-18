package routes

import (
	"go-api-starterkit/appsetup"
	"go-api-starterkit/config"

	"github.com/gin-gonic/gin"
	"go-api-starterkit/services"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB, cfg config.AppConfig, jwtService *services.JWTService) {
	appsetup.RegisterRoutes(router, db, cfg, jwtService)
}
