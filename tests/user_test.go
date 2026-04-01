package tests

import (
	"go-auth-app/controllers"
	"go-auth-app/middleware"
	"go-auth-app/services"
	"go-auth-app/tests/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAdminForbidden(t *testing.T) {
	mockRepo := &mocks.UserRepository{}

	userService := &services.UserService{
		UserRepo: mockRepo,
	}

	controller := controllers.UserController{
		UserService: userService,
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Middleware to simulate a logged-in user *before* admin middleware
	r.Use(func(c *gin.Context) {
		c.Set("role", "user")
		c.Next()
	})

	r.GET("/admin/users", middleware.AdminOnly(), func(c *gin.Context) {
		controller.GetAllUsers(c)
	})

	req, _ := http.NewRequest("GET", "/admin/users", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
}
