package auth

import (
	"pleco-api/internal/cache"
	"pleco-api/internal/config"
	"pleco-api/internal/erroroptimizer"
	"pleco-api/internal/modules/audit"
	"pleco-api/internal/modules/permission"
	userModule "pleco-api/internal/modules/user"
	"pleco-api/internal/services"

	"gorm.io/gorm"
)

type Module struct {
	Service AuthService
	Handler *AuthHandler
}

func BuildModule(
	db *gorm.DB,
	cfg config.AppConfig,
	userService *userModule.Service,
	jwtService *services.JWTService,
	auditService *audit.Service,
	permissionService *permission.Service,
	errorOptimizer *erroroptimizer.ErrorOptimizerService,
	stores ...cache.Store,
) *Module {
	service := NewService(db, cfg, userService, jwtService, auditService)
	if len(stores) > 0 {
		if impl, ok := service.(*authService); ok {
			impl.Cache = stores[0]
		}
	}
	handler := NewHandler(service, permissionService)
	handler.ErrorOptimizer = errorOptimizer
	if len(stores) > 0 {
		handler.Cache = stores[0]
	}

	return &Module{
		Service: service,
		Handler: handler,
	}
}
