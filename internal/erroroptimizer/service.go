package erroroptimizer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"pleco-api/internal/ai"
	"pleco-api/internal/cache"

	"gorm.io/gorm"
)

type ErrorOptimizerService struct {
	classifier ErrorClassifier
	aiService  *ai.Service
	cache      cache.Store
	db         *gorm.DB
	logger     *slog.Logger
}

func NewErrorOptimizerService(
	classifier ErrorClassifier,
	aiService *ai.Service,
	cache cache.Store,
	db *gorm.DB,
	logger *slog.Logger,
) *ErrorOptimizerService {
	return &ErrorOptimizerService{
		classifier: classifier,
		aiService:  aiService,
		cache:      cache,
		db:         db,
		logger:     logger,
	}
}

func (eos *ErrorOptimizerService) GetOptimizedError(
	ctx context.Context,
	err error,
	userContext UserContext,
	endpoint string,
) (*OptimizedError, error) {

	// 1. Classify error
	metadata := eos.classifier.Classify(err, endpoint)
	if metadata == nil {
		return nil, fmt.Errorf("failed to classify error")
	}

	// 2. Try cache
	if eos.cache != nil {
		cacheKey := fmt.Sprintf("error:%s:%s", metadata.Code, userContext.Language)
		var optimized OptimizedError
		if ok, err := eos.cache.GetJSON(ctx, cacheKey, &optimized); ok && err == nil {
			eos.logErrorOccurrence(ctx, metadata, userContext)
			return &optimized, nil
		}
	}

	// 3. Generate with AI
	var optimized *OptimizedError
	var aiErr error
	if eos.aiService != nil {
		optimized, aiErr = eos.generateWithAI(ctx, metadata, userContext, endpoint)
	} else {
		aiErr = fmt.Errorf("AI service not configured")
	}

	if aiErr != nil {
		if eos.logger != nil {
			eos.logger.Warn("failed to generate AI error message",
				slog.String("error_code", metadata.Code),
				slog.String("error", aiErr.Error()),
			)
		}
		// Fallback to generic message
		return eos.getFallbackError(metadata), nil
	}

	// 4. Cache result
	if eos.cache != nil {
		cacheKey := fmt.Sprintf("error:%s:%s", metadata.Code, userContext.Language)
		_ = eos.cache.SetJSON(ctx, cacheKey, optimized, 24*time.Hour)
	}

	// 5. Log
	eos.logErrorOccurrence(ctx, metadata, userContext)

	return optimized, nil
}

func (eos *ErrorOptimizerService) generateWithAI(
	ctx context.Context,
	metadata *ErrorMetadata,
	userContext UserContext,
	endpoint string,
) (*OptimizedError, error) {

	// Build prompt
	prompt := eos.buildPrompt(metadata, userContext, endpoint)

	// Call AI
	result, err := eos.aiService.Generate(ctx, ai.BuildJSONPrompt(
		"Generate a helpful, user-friendly error message response. Respond ONLY with valid JSON matching the exact schema requested.",
		prompt,
	))
	if err != nil {
		return nil, err
	}

	// Parse response
	var optimized OptimizedError
	if err := json.Unmarshal([]byte(result.Text), &optimized); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	if optimized.Message == "" {
		return nil, fmt.Errorf("AI response missing required 'message' field")
	}

	// Ensure code is set
	optimized.Code = metadata.Code

	return &optimized, nil
}

func (eos *ErrorOptimizerService) buildPrompt(
	metadata *ErrorMetadata,
	userContext UserContext,
	endpoint string,
) string {
	return fmt.Sprintf(`
Error Code: %s
Error Type: %s
Severity: %s
Endpoint: %s
User Language: %s
Is New User: %v
Recent Errors: %v

Generate a helpful error message following the JSON schema:
{
  "message": "User-friendly explanation of what went wrong",
  "details": "Additional context if helpful (optional, 1 sentence max)",
  "suggestions": [
    {
      "title": "Action title",
      "description": "Brief description",
      "action": "retry|navigate|contact_support|verify_email|reset_password",
      "url": "/path/or/url",
      "priority": "primary|secondary|tertiary"
    }
  ],
  "docs_url": "/docs/errors/%s"
}

Remember: be helpful but don't expose internal details.
`, metadata.Code, metadata.Type, metadata.Severity,
		endpoint, userContext.Language, userContext.IsNewUser,
		userContext.PreviousErrors, metadata.Code)
}

func (eos *ErrorOptimizerService) getFallbackError(metadata *ErrorMetadata) *OptimizedError {
	return &OptimizedError{
		Code:    metadata.Code,
		Message: metadata.UserMessage,
		Suggestions: []Suggestion{
			{
				Title:    "Try again",
				Action:   "retry",
				Priority: "primary",
			},
			{
				Title:    "Contact support",
				Action:   "contact_support",
				URL:      "/support",
				Priority: "secondary",
			},
		},
	}
}

func (eos *ErrorOptimizerService) logErrorOccurrence(
	ctx context.Context,
	metadata *ErrorMetadata,
	userContext UserContext,
) {
	if eos.db == nil {
		return
	}

	// This should ideally be asynchronous. Using a simple goroutine for now.
	go func() {
		// Log the occurrence in error_analytics
		err := eos.db.Exec(`
			INSERT INTO error_analytics (error_code, error_type, occurrence_count, last_occurred) 
			VALUES (?, ?, 1, NOW())
			ON CONFLICT (error_code) DO UPDATE 
			SET occurrence_count = error_analytics.occurrence_count + 1, last_occurred = NOW()
		`, metadata.Code, metadata.Type).Error

		if err != nil && eos.logger != nil {
			eos.logger.Error("failed to log error occurrence", slog.String("error", err.Error()))
		}
	}()
}
