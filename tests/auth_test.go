package tests

import (
	"go-auth-app/controllers"
	"go-auth-app/models"
	"go-auth-app/services"
	"go-auth-app/tests/mocks"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTest() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestRegister_Success(t *testing.T) {
	mockService := new(mocks.AuthService)

	controller := controllers.AuthController{
		AuthService: mockService,
	}

	body := `{
		"name": "test",
		"email": "test@mail.com",
		"password": "123456"
	}`

	mockService.
		On("Register", mock.Anything, "123456").
		Return(nil)

	c, w := setupTest()

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	controller.Register(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	mockService := new(mocks.AuthService)

	controller := controllers.AuthController{
		AuthService: mockService,
	}

	body := `{
		"email": "test@mail.com",
		"password": "123456"
	}`

	mockService.
		On("Login",
			"test@mail.com",
			"123456",
			"web",
			mock.Anything,
			mock.Anything,
		).
		Return(&services.AuthTokens{
			AccessToken:  "abc",
			RefreshToken: "xyz",
		}, nil)

	c, w := setupTest()

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Device-ID", "web")
	c.Request = req

	controller.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogout_Success(t *testing.T) {
	mockService := new(mocks.AuthService)

	controller := controllers.AuthController{
		AuthService: mockService,
	}

	mockService.
		On("Logout", uint(1), "web").
		Return(nil)

	c, w := setupTest()

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("X-Device-ID", "web")
	c.Request = req

	// ✅ penting: key + type harus match utils
	c.Set("user_id", uint(1))

	controller.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRefreshToken_Success(t *testing.T) {
	mockService := new(mocks.AuthService)

	controller := controllers.AuthController{
		AuthService: mockService,
	}

	body := `{"refresh_token":"abc"}`

	mockService.
		On("RefreshToken", "abc").
		Return(&services.AuthTokens{
			AccessToken:  "new",
			RefreshToken: "refresh",
		}, nil)

	c, w := setupTest()

	req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	controller.RefreshToken(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProfile_Success(t *testing.T) {
	mockService := new(mocks.AuthService)

	controller := controllers.AuthController{
		AuthService: mockService,
	}

	mockService.
		On("GetProfile", uint(1)).
		Return(&models.User{
			ID:    1,
			Name:  "Test",
			Email: "test@mail.com",
			Role:  "user",
		}, nil)

	c, w := setupTest()

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	c.Request = req

	c.Set("user_id", uint(1)) // ✅ penting

	controller.Profile(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVerifyEmail_Success(t *testing.T) {
	mockService := new(mocks.AuthService)

	controller := controllers.AuthController{
		AuthService: mockService,
	}

	mockService.
		On("VerifyEmail", "token123").
		Return(nil)

	c, w := setupTest()

	req := httptest.NewRequest(http.MethodGet, "/verify?token=token123", nil)
	c.Request = req

	controller.VerifyEmail(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResendVerification_Success(t *testing.T) {
	mockService := new(mocks.AuthService)

	controller := controllers.AuthController{
		AuthService: mockService,
	}

	body := `{"email":"test@mail.com"}`

	mockService.
		On("ResendVerification", "test@mail.com").
		Return(nil)

	c, w := setupTest()

	req := httptest.NewRequest(http.MethodPost, "/resend", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	controller.ResendVerification(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
