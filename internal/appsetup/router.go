package appsetup

import (
	"pleco-api/internal/ai"
	"pleco-api/internal/config"
	"pleco-api/internal/httpx"
	"pleco-api/internal/middleware"
	"pleco-api/internal/modules/audit"
	"pleco-api/internal/modules/auth"
	"pleco-api/internal/modules/permission"
	"pleco-api/internal/modules/role"
	"pleco-api/internal/modules/user"
	"pleco-api/internal/services"

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
	authModule := auth.BuildModule(db, cfg, userModule.Service, jwtService, auditModule.Service, permissionModule.Service)

	rateStore := newRateLimitStore(cfg.RedisURL)
	tokenVersionSrc := accessTokenVersionAdapter{repo: userModule.Repository}
	auth.SetupRoutes(api, authModule.Handler, jwtService, rateStore)
	user.SetupRoutes(api, userModule.Handler, jwtService, permissionModule.Service, tokenVersionSrc)
	audit.SetupRoutes(api, auditModule.Handler, jwtService, permissionModule.Service, tokenVersionSrc)
	role.SetupRoutes(api, roleModule.Handler, jwtService, permissionModule.Service, tokenVersionSrc)
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
