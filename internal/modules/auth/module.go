package auth

import (
	"pleco-api/internal/config"
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

func BuildModule(db *gorm.DB, cfg config.AppConfig, userService *userModule.Service, jwtService *services.JWTService, auditService *audit.Service, permissionService *permission.Service) *Module {
	service := NewService(db, cfg, userService, jwtService, auditService)
	handler := NewHandler(service, permissionService)

	return &Module{
		Service: service,
		Handler: handler,
	}
}
