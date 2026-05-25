package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
)

func TestBodySizeLimitReturnsTooLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(BodySizeLimit(4))
	router.POST("/echo", func(c *gin.Context) {
		var body map[string]string
		if err := c.ShouldBindJSON(&body); err != nil {
			httpx.ValidationError(c, httpx.FormatValidationError(err))
			return
		}
		httpx.Success(c, http.StatusOK, "ok", body, nil)
	})

	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"name":"too-large"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusRequestEntityTooLarge, w.Code, w.Body.String())
	}
}
