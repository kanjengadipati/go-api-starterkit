package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestParsePaginationUsesLimitOffset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/items?limit=10&offset=30", nil)

	pagination := ParsePagination(c)

	if pagination.Limit != 10 || pagination.Offset != 30 || pagination.Page() != 4 {
		t.Fatalf("unexpected pagination: %+v page=%d", pagination, pagination.Page())
	}
}

func TestParsePaginationCapsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/items?limit=999&offset=-1", nil)

	pagination := ParsePagination(c)

	if pagination.Limit != MaxLimit || pagination.Offset != 0 {
		t.Fatalf("unexpected pagination: %+v", pagination)
	}
}
