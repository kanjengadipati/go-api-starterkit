package tests

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pleco-api/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestStructuredLogger_EmitsRequestIDAndStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	originalOutput := log.Writer()
	originalFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer log.SetOutput(originalOutput)
	defer log.SetFlags(originalFlags)

	router := gin.New()
	router.Use(middleware.RequestID())
	router.Use(middleware.StructuredLogger())
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/health?from=test", nil)
	req.Header.Set("X-Request-ID", "req-logger-123")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		t.Fatal("expected structured log output, got empty buffer")
	}

	rawJSON := strings.TrimSpace(strings.TrimPrefix(lines[len(lines)-1], logPrefix(lines[len(lines)-1])))

	var payload map[string]any
	if err := json.Unmarshal([]byte(rawJSON), &payload); err != nil {
		t.Fatalf("expected valid JSON log, got %q (%v)", rawJSON, err)
	}

	assert.Equal(t, "req-logger-123", payload["request_id"])
	assert.Equal(t, "GET", payload["method"])
	assert.Equal(t, "/health?from=test", payload["path"])
	assert.Equal(t, float64(http.StatusNoContent), payload["status"])
	assert.Equal(t, "request completed", payload["message"])
}

func logPrefix(line string) string {
	if idx := strings.Index(line, "{"); idx >= 0 {
		return line[:idx]
	}
	return ""
}
