package tests

import (
	"errors"
	auth "go-auth-app/modules/auth"
	social "go-auth-app/modules/social"
	token "go-auth-app/modules/token"
	user "go-auth-app/modules/user"
	"go-auth-app/tests/mocks"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type stubRefreshTokenRepo struct {
	findByUserAndDevice func(userID uint, deviceID string) (*token.RefreshToken, error)
	deleteByID          func(id uint) error
}

func (s *stubRefreshTokenRepo) Save(_ *token.RefreshToken) error {
	return nil
}

func (s *stubRefreshTokenRepo) FindByUserAndDevice(userID uint, deviceID string) (*token.RefreshToken, error) {
	if s.findByUserAndDevice != nil {
		return s.findByUserAndDevice(userID, deviceID)
	}
	return nil, nil
}

func (s *stubRefreshTokenRepo) FindByUser(_ uint) ([]token.RefreshToken, error) {
	return nil, nil
}

func (s *stubRefreshTokenRepo) DeleteByID(id uint) error {
	if s.deleteByID != nil {
		return s.deleteByID(id)
	}
	return nil
}

func (s *stubRefreshTokenRepo) DeleteByUser(_ uint) error {
	return nil
}

type stubUserRepo struct{}

func (s *stubUserRepo) Create(_ *user.User) error                { return nil }
func (s *stubUserRepo) FindByEmail(_ string) (*user.User, error) { return nil, nil }
func (s *stubUserRepo) FindByID(_ uint) (*user.User, error)      { return nil, nil }
func (s *stubUserRepo) Update(_ *user.User) error                { return nil }
func (s *stubUserRepo) FindAll() ([]user.User, error)            { return nil, nil }
func (s *stubUserRepo) FindAllWithFilter(_, _ int, _, _ string) ([]user.User, int64, error) {
	return nil, 0, nil
}
func (s *stubUserRepo) Delete(_ uint) error { return nil }

type stubEmailVerificationRepo struct{}

func (s *stubEmailVerificationRepo) Create(_ *token.EmailVerificationToken) error {
	return nil
}
func (s *stubEmailVerificationRepo) FindByToken(_ string) (*token.EmailVerificationToken, error) {
	return nil, nil
}
func (s *stubEmailVerificationRepo) DeleteByID(_ uint) error     { return nil }
func (s *stubEmailVerificationRepo) DeleteByUserID(_ uint) error { return nil }

type stubSocialRepo struct{}

func (s *stubSocialRepo) Create(_ *social.SocialAccount) error { return nil }
func (s *stubSocialRepo) FindByProvider(_, _ string) (*social.SocialAccount, error) {
	return nil, nil
}

func setupTest() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestRegister_Success(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := auth.AuthHandler{AuthService: mockService}

	body := `{"name":"test","email":"test@mail.com","password":"123456"}`

	mockService.On("Register", mock.Anything, "123456").Return(nil)

	c, w := setupTest()
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.Register(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := auth.AuthHandler{AuthService: mockService}

	body := `{"email":"test@mail.com","password":"123456"}`

	mockService.On(
		"Login",
		"test@mail.com",
		"123456",
		"web",
		mock.Anything,
		mock.Anything,
	).Return(&auth.AuthTokens{
		AccessToken:  "abc",
		RefreshToken: "xyz",
	}, nil)

	c, w := setupTest()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Device-ID", "web")
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogout_Success(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := auth.AuthHandler{AuthService: mockService}

	mockService.On("Logout", uint(1), "web").Return(nil)

	c, w := setupTest()
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("X-Device-ID", "web")
	c.Request = req
	c.Set("user_id", uint(1))

	handler.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthService_Logout_MissingTokenIsIgnored(t *testing.T) {
	refreshRepo := &stubRefreshTokenRepo{
		findByUserAndDevice: func(userID uint, deviceID string) (*token.RefreshToken, error) {
			assert.Equal(t, uint(1), userID)
			assert.Equal(t, "web", deviceID)
			return nil, gorm.ErrRecordNotFound
		},
		deleteByID: func(id uint) error {
			t.Fatalf("DeleteByID should not be called when token is missing, got id=%d", id)
			return nil
		},
	}

	service := auth.NewAuthService(
		&stubUserRepo{},
		refreshRepo,
		&stubEmailVerificationRepo{},
		&stubSocialRepo{},
		nil,
		nil,
	)

	err := service.Logout(1, "web")

	assert.NoError(t, err)
}

func TestAuthService_Logout_DeletesFoundToken(t *testing.T) {
	var deletedID uint

	refreshRepo := &stubRefreshTokenRepo{
		findByUserAndDevice: func(userID uint, deviceID string) (*token.RefreshToken, error) {
			assert.Equal(t, uint(1), userID)
			assert.Equal(t, "web", deviceID)
			return &token.RefreshToken{
				Model:     gorm.Model{ID: 9},
				UserID:    1,
				DeviceID:  "web",
				ExpiredAt: time.Now().Add(time.Hour),
			}, nil
		},
		deleteByID: func(id uint) error {
			deletedID = id
			return nil
		},
	}

	service := auth.NewAuthService(
		&stubUserRepo{},
		refreshRepo,
		&stubEmailVerificationRepo{},
		&stubSocialRepo{},
		nil,
		nil,
	)

	err := service.Logout(1, "web")

	assert.NoError(t, err)
	assert.Equal(t, uint(9), deletedID)
}

func TestRefreshToken_Success(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := auth.AuthHandler{AuthService: mockService}

	mockService.On("RefreshToken", "abc").Return(&auth.AuthTokens{
		AccessToken:  "new",
		RefreshToken: "refresh",
	}, nil)

	c, w := setupTest()
	req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(`{"refresh_token":"abc"}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.RefreshToken(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProfile_Success(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := auth.AuthHandler{AuthService: mockService}

	mockService.On("GetProfile", uint(1)).Return(&user.User{
		Model: gorm.Model{ID: 1},
		Name:  "Test",
		Email: "test@mail.com",
		Role:  "user",
	}, nil)

	c, w := setupTest()
	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	c.Request = req
	c.Set("user_id", uint(1))

	handler.Profile(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVerifyEmail_Success(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := auth.AuthHandler{AuthService: mockService}

	mockService.On("VerifyEmail", "token123").Return(nil)

	c, w := setupTest()
	req := httptest.NewRequest(http.MethodGet, "/verify?token=token123", nil)
	c.Request = req

	handler.VerifyEmail(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResendVerification_Success(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := auth.AuthHandler{AuthService: mockService}

	mockService.On("ResendVerification", "test@mail.com").Return(nil)

	c, w := setupTest()
	req := httptest.NewRequest(http.MethodPost, "/resend", strings.NewReader(`{"email":"test@mail.com"}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.ResendVerification(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestForgotPassword_Success(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := auth.AuthHandler{AuthService: mockService}

	mockService.On("ForgotPassword", "test@mail.com").Return(nil)

	c, w := setupTest()
	req := httptest.NewRequest(http.MethodPost, "/forgot", strings.NewReader(`{"email":"test@mail.com"}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.ForgotPassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResetPassword_Success(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := auth.AuthHandler{AuthService: mockService}

	mockService.On("ResetPassword", "token123", "newSecret123").Return(nil)

	c, w := setupTest()
	req := httptest.NewRequest(http.MethodPost, "/reset", strings.NewReader(`{"token":"token123","new_password":"newSecret123"}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.ResetPassword(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSocialLogin_Success(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := auth.AuthHandler{AuthService: mockService}

	body := `{"provider":"google","id_token":"validGoogleIdToken","device_id":"browser1","user_agent":"Mozilla/5.0","ip_address":"127.0.0.1"}`

	mockService.On(
		"SocialLogin",
		"google",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(&auth.AuthTokens{
		AccessToken:  "access_abc",
		RefreshToken: "refresh_xyz",
	}, nil)

	c, w := setupTest()
	req := httptest.NewRequest(http.MethodPost, "/social-login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.SocialLogin(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestSocialLogin_Failure(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := auth.AuthHandler{AuthService: mockService}

	body := `{"provider":"google","id_token":"invalidToken","device_id":"browser1","user_agent":"Mozilla/5.0","ip_address":"127.0.0.1"}`

	mockService.On(
		"SocialLogin",
		"google",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, errors.New("invalid google token"))

	c, w := setupTest()
	req := httptest.NewRequest(http.MethodPost, "/social-login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.SocialLogin(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}
