package httpx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pleco-api/internal/domain"

	"github.com/gin-gonic/gin"
)

func TestHandleErrorUsesAPIErrorEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	HandleError(c, domain.NewAPIError(http.StatusConflict, domain.CodeConflict, "already exists", errors.New("duplicate")))

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
	if body := w.Body.String(); body == "" || !containsAll(body, `"code":"CONFLICT"`, `"message":"already exists"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !contains(value, part) {
			return false
		}
	}
	return true
}

func contains(value, part string) bool {
	return strings.Contains(value, part)
}
