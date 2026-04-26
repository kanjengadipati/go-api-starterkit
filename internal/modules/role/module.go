package role

import (
	permissionModule "pleco-api/internal/modules/permission"

	"gorm.io/gorm"
)

type Module struct {
	Repository Repository
	Service    *Service
	Handler    *Handler
}

func BuildModule(db *gorm.DB) *Module {
	repository := NewRepository(db)
	permissionRepo := permissionModule.NewRepository(db)
	service := NewService(repository, permissionRepo)
	handler := NewHandler(service)

	return &Module{
		Repository: repository,
		Service:    service,
		Handler:    handler,
	}
}
