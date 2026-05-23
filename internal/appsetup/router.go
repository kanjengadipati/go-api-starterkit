package appsetup

import (
	"log/slog"
	"pleco-api/internal/ai"
	"pleco-api/internal/config"
	"pleco-api/internal/erroroptimizer"
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

func RegisterRoutes(router *gin.Engine, db *gorm.DB, cfg config.AppConfig, jwtService *services.JWTService, rateStore middleware.RateLimitStore) error {
	api := router.Group("/")
	cacheStore := newCacheStore(cfg)
	aiService, err := ai.NewService(cfg.AI)
	if err != nil {
		return err
	}
	permissionModule := permission.BuildModule(db, cacheStore)
	roleModule := role.BuildModule(db, cacheStore)
	auditModule := audit.BuildModule(db, aiService)
	userModule := user.BuildModule(db, auditModule.Service, cacheStore)
	userModule.Handler.PermissionSvc = permissionModule.Service
	userModule.Handler.Cache = cacheStore

	var errOptSvc *erroroptimizer.ErrorOptimizerService
	if cfg.ErrorOptimization.Enabled {
		classifier := &erroroptimizer.DefaultErrorClassifier{}
		errOptSvc = erroroptimizer.NewErrorOptimizerService(classifier, aiService, cacheStore, db, slog.Default())
	}

	authModule := auth.BuildModule(db, cfg, userModule.Service, jwtService, auditModule.Service, permissionModule.Service, errOptSvc, cacheStore)

	tokenVersionSrc := accessTokenVersionAdapter{repo: userModule.Repository}
	auth.SetupRoutes(api, authModule.Handler, jwtService, rateStore, tokenVersionSrc, cfg)
	user.SetupRoutes(api, userModule.Handler, jwtService, permissionModule.Service, tokenVersionSrc)
	audit.SetupRoutes(api, auditModule.Handler, jwtService, permissionModule.Service, tokenVersionSrc)
	role.SetupRoutes(api, roleModule.Handler, jwtService, permissionModule.Service, tokenVersionSrc)
	router.GET("/health", func(c *gin.Context) {
		httpx.Success(c, 200, "Health check ok", gin.H{"status": "ok"}, nil)
	})

	router.GET("/health/live", func(c *gin.Context) {
		httpx.Success(c, 200, "Service is live", gin.H{"status": "ok"}, nil)
	})

	router.GET("/health/ready", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			httpx.Error(c, 503, "Database connection error")
			return
		}

		if err := sqlDB.Ping(); err != nil {
			httpx.Error(c, 503, "Database ping failed")
			return
		}

		httpx.Success(c, 200, "Service is ready", gin.H{"status": "ok"}, nil)
	})
	return nil
}

func BuildRouter(db *gorm.DB, cfg config.AppConfig, jwtService *services.JWTService, rateStore middleware.RateLimitStore) (*gin.Engine, error) {
	router := gin.New()
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, err
	}
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))
	router.Use(middleware.RequestID())
	router.Use(middleware.StructuredLogger())
	router.Use(middleware.RecoveryLogger())
	router.Use(middleware.SecurityHeaders())
	if err := RegisterRoutes(router, db, cfg, jwtService, rateStore); err != nil {
		return nil, err
	}
	return router, nil
}
