package appsetup

import (
	"go-api-starterkit/internal/ai"
	"go-api-starterkit/internal/config"
	"go-api-starterkit/internal/httpx"
	"go-api-starterkit/internal/middleware"
	"go-api-starterkit/internal/modules/audit"
	"go-api-starterkit/internal/modules/auth"
	"go-api-starterkit/internal/modules/permission"
	"go-api-starterkit/internal/modules/role"
	"go-api-starterkit/internal/modules/user"
	"go-api-starterkit/internal/services"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB, cfg config.AppConfig, jwtService *services.JWTService) error {
	api := router.Group("/")
	aiService, err := ai.NewService(cfg.AI)
	if err != nil {
		return err
	}
	permissionModule := permission.BuildModule(db)
	roleModule := role.BuildModule(db)
	auditModule := audit.BuildModule(db, aiService)
	userModule := user.BuildModule(db, auditModule.Service)
	authModule := auth.BuildModule(db, cfg, userModule.Service, jwtService, auditModule.Service)

	auth.SetupRoutes(api, authModule.Handler, jwtService)
	user.SetupRoutes(api, userModule.Handler, jwtService, permissionModule.Service)
	audit.SetupRoutes(api, auditModule.Handler, jwtService, permissionModule.Service)
	role.SetupRoutes(api, roleModule.Handler, jwtService, permissionModule.Service)
	router.GET("/health", func(c *gin.Context) {
		httpx.Success(c, 200, "Health check ok", gin.H{"status": "ok"}, nil)
	})
	return nil
}

func BuildRouter(db *gorm.DB, cfg config.AppConfig, jwtService *services.JWTService) (*gin.Engine, error) {
	router := gin.New()
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, err
	}
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))
	router.Use(middleware.RequestID())
	router.Use(middleware.StructuredLogger())
	router.Use(middleware.RecoveryLogger())
	router.Use(middleware.SecurityHeaders())
	if err := RegisterRoutes(router, db, cfg, jwtService); err != nil {
		return nil, err
	}
	return router, nil
}
