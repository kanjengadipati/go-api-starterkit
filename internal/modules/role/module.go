package role

import (
	"pleco-api/internal/cache"
	"pleco-api/internal/modules/audit"
	permissionModule "pleco-api/internal/modules/permission"

	"gorm.io/gorm"
)

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func BuildModule(db *gorm.DB, auditSvc *audit.Service, stores ...cache.Store) *Module {
	repository := NewRepository(db)
	permissionRepo := permissionModule.NewRepository(db)
	service := NewService(repository, permissionRepo)
	service.AuditSvc = auditSvc
	if len(stores) > 0 {
		service.Cache = stores[0]
	}
	handler := NewHandler(service)

	return &Module{
		Repository: repository,
		Service:    service,
		Handler:    handler,
	}
}
