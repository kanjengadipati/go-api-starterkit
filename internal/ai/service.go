package ai

import (
	"context"
	"errors"
	"time"

	"pleco-api/internal/config"
)

var ErrDisabled = errors.New("ai is disabled")
var ErrTimeout = errors.New("ai request timed out")
var ErrInvalidStructuredOutput = errors.New("ai returned an invalid structured response")

type Service struct {
	enabled      bool
	model        string
	providerName string
	provider     Provider
}

func NewService(cfg config.AIConfig) (*Service, error) {
	service := &Service{
		enabled:      cfg.Enabled,
		model:        cfg.Model,
		providerName: cfg.Provider,
	}

	if !cfg.Enabled {
		return service, nil
	}

	switch cfg.Provider {
	case "mock":
		service.provider = NewMockProvider()
	case "ollama":
		service.provider = NewOllamaProvider(cfg.BaseURL, time.Duration(cfg.TimeoutSeconds)*time.Second)
	case "openai":
		service.provider = NewOpenAIProvider(cfg.BaseURL, cfg.APIKey, time.Duration(cfg.TimeoutSeconds)*time.Second)
	case "gemini":
		service.provider = NewGeminiProvider(cfg.BaseURL, cfg.APIKey, time.Duration(cfg.TimeoutSeconds)*time.Second)
	default:
		return nil, errors.New("unsupported ai provider")
	}

	return service, nil
}

func (s *Service) Enabled() bool {
	return s != nil && s.enabled && s.provider != nil
}

func (s *Service) Generate(ctx context.Context, input GenerateInput) (*GenerateResult, error) {
	if s == nil || !s.enabled || s.provider == nil {
		return nil, ErrDisabled
	}
	if input.Model == "" {
		input.Model = s.model
	}
	return s.provider.Generate(ctx, input)
}

func (s *Service) ProviderName() string {
	if s == nil {
		return ""
	}
	return s.providerName
}

func (s *Service) ModelName() string {
	if s == nil {
		return ""
	}
	return s.model
}
