package audit

import (
	"pleco-api/internal/ai"

	"gorm.io/gorm"
)

type Module struct {
	Repository Repository
	Service    *Service
	AIService  *InvestigatorService
	Handler    *Handler
}

func BuildModule(db *gorm.DB, aiService *ai.Service) *Module {
	repository := NewRepository(db)
	service := NewService(repository)
	investigatorService := NewInvestigatorService(repository, aiService, service)
	handler := NewHandler(service, investigatorService)

	return &Module{
		Repository: repository,
		Service:    service,
		AIService:  investigatorService,
		Handler:    handler,
	}
}
