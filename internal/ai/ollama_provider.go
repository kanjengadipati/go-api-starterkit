package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OllamaProvider struct {
	baseURL string
	client  *http.Client
}

type ollamaRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	System      string  `json:"system,omitempty"`
	Stream      bool    `json:"stream"`
	Temperature float64 `json:"temperature,omitempty"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

var (
	ErrOllamaUnavailable  = errors.New("ollama is unavailable")
	ErrOllamaModelMissing = errors.New("ollama model is not available")
)

func NewOllamaProvider(baseURL string, timeout time.Duration) *OllamaProvider {
	return &OllamaProvider{
		baseURL: normalizeBaseURL(baseURL),
		client:  &http.Client{Timeout: timeout},
	}
}

func (p *OllamaProvider) Generate(ctx context.Context, input GenerateInput) (*GenerateResult, error) {
	body, err := json.Marshal(ollamaRequest{
		Model:       input.Model,
		Prompt:      input.UserPrompt,
		System:      input.SystemPrompt,
		Stream:      false,
		Temperature: input.Temperature,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: %v", ErrTimeout, err)
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return nil, fmt.Errorf("%w: %v", ErrTimeout, err)
		}
		return nil, fmt.Errorf("%w: %v", ErrOllamaUnavailable, err)
	}
	defer resp.Body.Close()

	var parsed ollamaResponse
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return nil, fmt.Errorf("failed to decode ollama response: %s", strings.TrimSpace(string(bodyBytes)))
	}

	if resp.StatusCode >= 400 {
		if resp.StatusCode == http.StatusNotFound || strings.Contains(strings.ToLower(parsed.Error), "model") {
			return nil, fmt.Errorf("%w: %s", ErrOllamaModelMissing, fallbackOllamaError(parsed.Error, resp.StatusCode))
		}
		if parsed.Error != "" {
			return nil, fmt.Errorf("ollama error: %s", parsed.Error)
		}
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}
	if parsed.Error != "" {
		return nil, fmt.Errorf("ollama error: %s", parsed.Error)
	}

	return &GenerateResult{Text: parsed.Response}, nil
}

func (p *OllamaProvider) HealthCheck(ctx context.Context) error {
	endpoint := p.baseURL + "/api/tags"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrOllamaUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ollama health check returned status %d", resp.StatusCode)
	}

	return nil
}

func fallbackOllamaError(message string, status int) string {
	if strings.TrimSpace(message) != "" {
		return message
	}
	return fmt.Sprintf("status %d", status)
}

func normalizeBaseURL(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return strings.TrimRight(raw, "/")
	}
	return strings.TrimRight(parsed.String(), "/")
}
