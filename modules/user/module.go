package user

import (
	"go-api-starterkit/modules/audit"

	"gorm.io/gorm"
)

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func BuildModule(db *gorm.DB, auditSvc *audit.Service) *Module {
	repository := NewRepository(db)
	service := NewService(repository, auditSvc)
	handler := NewHandler(service)

	return &Module{
		Repository: repository,
		Service:    service,
		Handler:    handler,
	}
}
