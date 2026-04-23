package ai

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestOllamaProviderGenerateSuccess(t *testing.T) {
	provider := NewOllamaProvider("http://ollama.local", 5*time.Second)
	provider.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Equal(t, "/api/generate", req.URL.Path)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"response":"{\"summary\":\"ok\",\"timeline\":[],\"suspicious_signals\":[],\"recommendations\":[]}","error":""}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	result, err := provider.Generate(context.Background(), GenerateInput{
		Model:      "qwen2.5:3b",
		UserPrompt: "hello",
	})

	assert.NoError(t, err)
	assert.Contains(t, result.Text, `"summary":"ok"`)
}

func TestOllamaProviderGenerateReturnsModelMissing(t *testing.T) {
	provider := NewOllamaProvider("http://ollama.local", 5*time.Second)
	provider.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader(`{"error":"model 'qwen2.5:3b' not found"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	result, err := provider.Generate(context.Background(), GenerateInput{
		Model:      "qwen2.5:3b",
		UserPrompt: "hello",
	})

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOllamaModelMissing)
}

func TestOllamaProviderGenerateReturnsTimeout(t *testing.T) {
	provider := NewOllamaProvider("http://ollama.local", 5*time.Second)
	provider.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, context.DeadlineExceeded
		}),
	}

	result, err := provider.Generate(context.Background(), GenerateInput{
		Model:      "qwen2.5:3b",
		UserPrompt: "hello",
	})

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrTimeout))
}
