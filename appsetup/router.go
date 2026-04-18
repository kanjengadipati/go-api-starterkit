package appsetup

import (
	"go-api-starterkit/config"
	"go-api-starterkit/httpx"
	"go-api-starterkit/modules/audit"
	"go-api-starterkit/modules/auth"
	"go-api-starterkit/modules/permission"
	"go-api-starterkit/modules/user"
	"go-api-starterkit/services"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB, cfg config.AppConfig, jwtService *services.JWTService) {
	api := router.Group("/")
	permissionModule := permission.BuildModule(db)
	auditModule := audit.BuildModule(db)
	userModule := user.BuildModule(db, auditModule.Service)
	authModule := auth.BuildModule(db, cfg, userModule.Service, jwtService, auditModule.Service)

	auth.SetupRoutes(api, authModule.Handler, jwtService)
	user.SetupRoutes(api, userModule.Handler, jwtService, permissionModule.Service)
	audit.SetupRoutes(api, auditModule.Handler, jwtService, permissionModule.Service)
	router.GET("/health", func(c *gin.Context) {
		httpx.Success(c, 200, "Health check ok", gin.H{"status": "ok"}, nil)
	})
}

func BuildRouter(db *gorm.DB, cfg config.AppConfig, jwtService *services.JWTService) *gin.Engine {
	router := gin.Default()
	RegisterRoutes(router, db, cfg, jwtService)
	return router
}
