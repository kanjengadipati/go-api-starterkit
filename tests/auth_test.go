package tests

import (
	"encoding/hex"
	"errors"
	"go-api-starterkit/internal/appsetup"
	"go-api-starterkit/internal/config"
	"go-api-starterkit/internal/middleware"
	auth "go-api-starterkit/internal/modules/auth"
	"go-api-starterkit/internal/modules/permission"
	social "go-api-starterkit/internal/modules/social"
	token "go-api-starterkit/internal/modules/token"
	user "go-api-starterkit/internal/modules/user"
	"go-api-starterkit/internal/services"
	"go-api-starterkit/tests/mocks"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type stubRefreshTokenRepo struct {
	findByID            func(id uint) (*token.RefreshToken, error)
	findByUserAndDevice func(userID uint, deviceID string) (*token.RefreshToken, error)
	findByTokenHash     func(tokenHash string) (*token.RefreshToken, error)
	findByUser          func(userID uint) ([]token.RefreshToken, error)
	deleteByID          func(id uint) error
	deleteByUserAndID   func(userID, id uint) error
	deleteByUser        func(userID uint) error
	deleteByUserExcept  func(userID uint, deviceID string) error
}

func (s *stubRefreshTokenRepo) Save(_ *token.RefreshToken) error {
	return nil
}

func (s *stubRefreshTokenRepo) FindByID(id uint) (*token.RefreshToken, error) {
	if s.findByID != nil {
		return s.findByID(id)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubRefreshTokenRepo) FindByUserAndDevice(userID uint, deviceID string) (*token.RefreshToken, error) {
	if s.findByUserAndDevice != nil {
		return s.findByUserAndDevice(userID, deviceID)
	}
	return nil, nil
}

func (s *stubRefreshTokenRepo) FindByTokenHash(tokenHash string) (*token.RefreshToken, error) {
	if s.findByTokenHash != nil {
		return s.findByTokenHash(tokenHash)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubRefreshTokenRepo) FindByUser(userID uint) ([]token.RefreshToken, error) {
	if s.findByUser != nil {
		return s.findByUser(userID)
	}
	return nil, nil
}

func (s *stubRefreshTokenRepo) DeleteByID(id uint) error {
	if s.deleteByID != nil {
		return s.deleteByID(id)
	}
	return nil
}

func (s *stubRefreshTokenRepo) DeleteByUserAndID(userID, id uint) error {
	if s.deleteByUserAndID != nil {
		return s.deleteByUserAndID(userID, id)
	}
	return nil
}

func (s *stubRefreshTokenRepo) DeleteByUser(userID uint) error {
	if s.deleteByUser != nil {
		return s.deleteByUser(userID)
	}
	return nil
}

func (s *stubRefreshTokenRepo) DeleteByUserExceptDevice(userID uint, deviceID string) error {
	if s.deleteByUserExcept != nil {
		return s.deleteByUserExcept(userID, deviceID)
	}
	return nil
}
func (s *stubRefreshTokenRepo) WithTx(_ *gorm.DB) token.RefreshTokenRepository {
	return s
}

type stubUserRepo struct {
	create      func(*user.User) error
	findByEmail func(string) (*user.User, error)
	findByID    func(uint) (*user.User, error)
	update      func(*user.User) error
}

func (s *stubUserRepo) Create(u *user.User) error {
	if s.create != nil {
		return s.create(u)
	}
	return nil
}
func (s *stubUserRepo) FindByEmail(email string) (*user.User, error) {
	if s.findByEmail != nil {
		return s.findByEmail(email)
	}
	return nil, nil
}
func (s *stubUserRepo) FindByID(id uint) (*user.User, error) {
	if s.findByID != nil {
		return s.findByID(id)
	}
	return nil, nil
}
func (s *stubUserRepo) Update(u *user.User) error {
	if s.update != nil {
		return s.update(u)
	}
	return nil
}
func (s *stubUserRepo) WithTx(_ *gorm.DB) user.Repository {
	return s
}
func (s *stubUserRepo) FindAll() ([]user.User, error) { return nil, nil }
func (s *stubUserRepo) FindAllWithFilter(_, _ int, _, _ string) ([]user.User, int64, error) {
	return nil, 0, nil
}
func (s *stubUserRepo) Delete(_ uint) error { return nil }

type stubEmailVerificationRepo struct {
	create         func(*token.EmailVerificationToken) error
	findByToken    func(string) (*token.EmailVerificationToken, error)
	deleteByID     func(uint) error
	deleteByUserID func(uint) error
}

func (s *stubEmailVerificationRepo) Create(tk *token.EmailVerificationToken) error {
	if s.create != nil {
		return s.create(tk)
	}
	return nil
}
func (s *stubEmailVerificationRepo) FindByToken(value string) (*token.EmailVerificationToken, error) {
	if s.findByToken != nil {
		return s.findByToken(value)
	}
	return nil, nil
}
func (s *stubEmailVerificationRepo) DeleteByID(id uint) error {
	if s.deleteByID != nil {
		return s.deleteByID(id)
	}
	return nil
}
func (s *stubEmailVerificationRepo) DeleteByUserID(userID uint) error {
	if s.deleteByUserID != nil {
		return s.deleteByUserID(userID)
	}
	return nil
}
func (s *stubEmailVerificationRepo) WithTx(_ *gorm.DB) token.EmailVerificationRepository {
	return s
}

type stubSocialRepo struct{}

func (s *stubSocialRepo) Create(_ *social.SocialAccount) error { return nil }
func (s *stubSocialRepo) FindByProvider(_, _ string) (*social.SocialAccount, error) {
	return nil, nil
}
func (s *stubSocialRepo) WithTx(_ *gorm.DB) social.Repository {
	return s
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

	body := `{"name":"test","email":"test@mail.com","password":"12345678"}`

	mockService.On("Register", mock.Anything, "12345678").Return(nil)

	c, w := setupTest()
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.Register(c)

	assert.Equal(t, http.StatusOK, w.Code)
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "success", bodyMap["status"])
	assert.Equal(t, "User registered", bodyMap["message"])
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
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "success", bodyMap["status"])
	assert.Equal(t, "Login success", bodyMap["message"])
	data, ok := bodyMap["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "abc", data["access_token"])
	assert.Equal(t, "xyz", data["refresh_token"])
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
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "success", bodyMap["status"])
	assert.Equal(t, "logout success", bodyMap["message"])
}

func TestAuthMiddleware_RejectsRefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	jwtService := services.NewJWTService([]byte("super_secret_key_123_must_be_32_bytes_long_minimum"))

	router.GET("/protected", middleware.AuthMiddleware(jwtService), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	refreshToken, err := jwtService.GenerateToken(1, "user", time.Minute, "refresh", 0)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "error", bodyMap["status"])
	assert.Equal(t, "invalid token", bodyMap["message"])
}

func TestAuthMiddleware_RejectsMalformedClaimsGracefully(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	jwtService := services.NewJWTService([]byte("super_secret_key_123_must_be_32_bytes_long_minimum"))

	router.GET("/protected", middleware.AuthMiddleware(jwtService), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	tokenString, err := jwtService.GenerateCustomClaimsToken(map[string]interface{}{
		"user_id": "not-a-number",
		"role":    "user",
		"type":    "access",
		"tv":      float64(0),
	}, time.Minute)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "error", bodyMap["status"])
	assert.Equal(t, "invalid token", bodyMap["message"])
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
		nil,
		&stubUserRepo{},
		refreshRepo,
		&stubEmailVerificationRepo{},
		&stubSocialRepo{},
		nil,
		nil,
		nil,
		config.SocialConfig{},
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
		nil,
		&stubUserRepo{},
		refreshRepo,
		&stubEmailVerificationRepo{},
		&stubSocialRepo{},
		nil,
		nil,
		nil,
		config.SocialConfig{},
	)

	err := service.Logout(1, "web")

	assert.NoError(t, err)
	assert.Equal(t, uint(9), deletedID)
}

func TestAuthService_Register_IgnoresEmailDeliveryFailure(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	userRepo := &stubUserRepo{}
	verificationRepo := &stubEmailVerificationRepo{}
	emailSvc := mocks.NewEmailService(t)

	createdUser := 0
	createdVerification := 0

	userRepo.create = func(u *user.User) error {
		createdUser++
		u.Model = gorm.Model{ID: 1}
		return nil
	}
	verificationRepo.create = func(tk *token.EmailVerificationToken) error {
		createdVerification++
		assert.Equal(t, uint(1), tk.UserID)
		assert.Len(t, tk.Token, 64, "verification token should be stored as SHA-256 hex")
		_, decErr := hex.DecodeString(tk.Token)
		assert.NoError(t, decErr)
		return nil
	}
	emailSvc.On("SendVerificationEmail", "test@mail.com", mock.Anything).Return(errors.New("sendgrid unavailable"))

	service := auth.NewAuthService(
		db,
		userRepo,
		&stubRefreshTokenRepo{},
		verificationRepo,
		&stubSocialRepo{},
		services.NewJWTService([]byte("super_secret_key_123_must_be_32_bytes_long_minimum")),
		emailSvc,
		nil,
		config.SocialConfig{},
	)

	err = service.Register(&user.User{
		Name:  "Test",
		Email: "test@mail.com",
	}, "12345678")

	assert.NoError(t, err)
	assert.Equal(t, 1, createdUser)
	assert.Equal(t, 1, createdVerification)
}

func TestAuthService_ListSessions_MarksCurrentDevice(t *testing.T) {
	refreshRepo := &stubRefreshTokenRepo{
		findByUser: func(userID uint) ([]token.RefreshToken, error) {
			assert.Equal(t, uint(7), userID)
			return []token.RefreshToken{
				{
					Model:     gorm.Model{ID: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					UserID:    7,
					DeviceID:  "phone",
					UserAgent: "Mobile",
					IPAddress: "10.0.0.2",
					ExpiredAt: time.Now().Add(time.Hour),
				},
				{
					Model:     gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
					UserID:    7,
					DeviceID:  "web",
					UserAgent: "Browser",
					IPAddress: "10.0.0.1",
					ExpiredAt: time.Now().Add(time.Hour),
				},
			}, nil
		},
	}

	service := auth.NewAuthService(
		nil,
		&stubUserRepo{},
		refreshRepo,
		&stubEmailVerificationRepo{},
		&stubSocialRepo{},
		nil,
		nil,
		nil,
		config.SocialConfig{},
	)

	sessions, err := service.ListSessions(7, "web")

	assert.NoError(t, err)
	assert.Len(t, sessions, 2)
	assert.Equal(t, uint(2), sessions[0].ID)
	assert.False(t, sessions[0].IsCurrent)
	assert.Equal(t, uint(1), sessions[1].ID)
	assert.True(t, sessions[1].IsCurrent)
}

func TestAuthService_LogoutOtherSessions(t *testing.T) {
	refreshRepo := &stubRefreshTokenRepo{
		deleteByUserExcept: func(userID uint, deviceID string) error {
			assert.Equal(t, uint(9), userID)
			assert.Equal(t, "web", deviceID)
			return nil
		},
	}

	service := auth.NewAuthService(
		nil,
		&stubUserRepo{},
		refreshRepo,
		&stubEmailVerificationRepo{},
		&stubSocialRepo{},
		nil,
		nil,
		nil,
		config.SocialConfig{},
	)

	err := service.LogoutOtherSessions(9, "web", "Browser", "127.0.0.1")

	assert.NoError(t, err)
}

func TestAuthService_RevokeSession_DeletesOwnedSession(t *testing.T) {
	var deletedSessionID uint

	refreshRepo := &stubRefreshTokenRepo{
		findByID: func(id uint) (*token.RefreshToken, error) {
			assert.Equal(t, uint(4), id)
			return &token.RefreshToken{
				Model:    gorm.Model{ID: 4},
				UserID:   11,
				DeviceID: "tablet",
			}, nil
		},
		deleteByUserAndID: func(userID, id uint) error {
			assert.Equal(t, uint(11), userID)
			deletedSessionID = id
			return nil
		},
	}

	service := auth.NewAuthService(
		nil,
		&stubUserRepo{},
		refreshRepo,
		&stubEmailVerificationRepo{},
		&stubSocialRepo{},
		nil,
		nil,
		nil,
		config.SocialConfig{},
	)

	err := service.RevokeSession(11, 4, "Browser", "127.0.0.1")

	assert.NoError(t, err)
	assert.Equal(t, uint(4), deletedSessionID)
}

func TestAuthService_RevokeSession_RejectsForeignSession(t *testing.T) {
	refreshRepo := &stubRefreshTokenRepo{
		findByID: func(id uint) (*token.RefreshToken, error) {
			return &token.RefreshToken{
				Model:  gorm.Model{ID: id},
				UserID: 99,
			}, nil
		},
	}

	service := auth.NewAuthService(
		nil,
		&stubUserRepo{},
		refreshRepo,
		&stubEmailVerificationRepo{},
		&stubSocialRepo{},
		nil,
		nil,
		nil,
		config.SocialConfig{},
	)

	err := service.RevokeSession(11, 4, "Browser", "127.0.0.1")

	assert.EqualError(t, err, "session not found")
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
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "success", bodyMap["status"])
	assert.Equal(t, "Refresh token success", bodyMap["message"])
	data, ok := bodyMap["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "new", data["access_token"])
	assert.Equal(t, "refresh", data["refresh_token"])
}

func TestProfile_Success(t *testing.T) {
	mockService := new(mocks.AuthService)
	permSvc := permission.NewService(&stubPermissionRepo{})
	handler := auth.AuthHandler{
		AuthService:   mockService,
		PermissionSvc: permSvc,
	}

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
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "success", bodyMap["status"])
	assert.Equal(t, "Profile fetched", bodyMap["message"])
	data, ok := bodyMap["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "test@mail.com", data["email"])
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
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "success", bodyMap["status"])
	assert.Equal(t, "email verified", bodyMap["message"])
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

	body := `{"provider":"google","token":"validGoogleIdToken","device_id":"browser1","user_agent":"Mozilla/5.0","ip_address":"127.0.0.1"}`

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

	body := `{"provider":"google","token":"invalidToken","device_id":"browser1","user_agent":"Mozilla/5.0","ip_address":"127.0.0.1"}`

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

func TestResetPassword_ValidationFailure(t *testing.T) {
	mockService := new(mocks.AuthService)
	handler := auth.AuthHandler{AuthService: mockService}

	body := `{"token":"token123","new_password":"short"}`

	c, w := setupTest()
	req := httptest.NewRequest(http.MethodPost, "/reset", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.ResetPassword(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "error", bodyMap["status"])
	assert.Contains(t, bodyMap["message"], "Validation failed")
	mockService.AssertNotCalled(t, "ResetPassword", mock.Anything, mock.Anything)
}

func TestAuthRateLimiter_ReturnsTooManyRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	limiter := middleware.NewRateLimiter(1, time.Minute)

	router.POST("/auth/login", limiter.Middleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	firstReq := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	firstReq.RemoteAddr = "203.0.113.10:1234"
	firstResp := httptest.NewRecorder()
	router.ServeHTTP(firstResp, firstReq)

	secondReq := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	secondReq.RemoteAddr = "203.0.113.10:1234"
	secondResp := httptest.NewRecorder()
	router.ServeHTTP(secondResp, secondReq)

	assert.Equal(t, http.StatusOK, firstResp.Code)
	assert.Equal(t, http.StatusTooManyRequests, secondResp.Code)
	bodyMap := decodeBodyMap(t, secondResp)
	assert.Equal(t, "error", bodyMap["status"])
	assert.Equal(t, "too many requests", bodyMap["message"])
}

func TestSecurityHeadersMiddleware_SetsExpectedHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.SecurityHeaders())
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, "nosniff", resp.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", resp.Header().Get("X-Frame-Options"))
	assert.Equal(t, "no-referrer", resp.Header().Get("Referrer-Policy"))
	assert.Equal(t, "none", resp.Header().Get("X-Permitted-Cross-Domain-Policies"))
	assert.Equal(t, "default-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'", resp.Header().Get("Content-Security-Policy"))
	assert.Empty(t, resp.Header().Get("Strict-Transport-Security"))
}

func TestSecurityHeadersMiddleware_SetsHSTSForHTTPSRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.SecurityHeaders())
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, "max-age=31536000; includeSubDomains", resp.Header().Get("Strict-Transport-Security"))
}

func TestRequestIDMiddleware_UsesIncomingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestID())
	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "req-123")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, "req-123", resp.Header().Get("X-Request-ID"))
}

func TestBuildRouter_RejectsInvalidTrustedProxy(t *testing.T) {
	cfg := config.AppConfig{
		Port:           "8080",
		DatabaseURL:    "postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable",
		TrustedProxies: []string{"definitely-not-a-cidr"},
		JWTSecret:      []byte("super_secret_key_123_must_be_32_bytes_long_minimum"),
	}

	router, err := appsetup.BuildRouter(nil, cfg, services.NewJWTService([]byte("super_secret_key_123_must_be_32_bytes_long_minimum")))

	assert.Nil(t, router)
	assert.Error(t, err)
}
