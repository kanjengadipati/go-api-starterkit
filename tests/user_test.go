package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"pleco-api/internal/middleware"
	user "pleco-api/internal/modules/user"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type stubUserRepository struct {
	create      func(*user.User) error
	findByEmail func(string) (*user.User, error)
	findByID    func(uint) (*user.User, error)
	update      func(*user.User) error
	findAll     func() ([]user.User, error)
	delete      func(uint) error
}

func (s *stubUserRepository) Create(u *user.User) error {
	if s.create != nil {
		return s.create(u)
	}
	return nil
}
func (s *stubUserRepository) FindByEmail(email string) (*user.User, error) {
	if s.findByEmail != nil {
		return s.findByEmail(email)
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepository) FindByID(id uint) (*user.User, error) {
	if s.findByID != nil {
		return s.findByID(id)
	}
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepository) Update(u *user.User) error {
	if s.update != nil {
		return s.update(u)
	}
	return nil
}
func (s *stubUserRepository) FindAll() ([]user.User, error) {
	if s.findAll != nil {
		return s.findAll()
	}
	return nil, nil
}
func (s *stubUserRepository) FindAllWithFilter(_, _ int, _, _ string) ([]user.User, int64, error) {
	return nil, 0, nil
}
func (s *stubUserRepository) Delete(id uint) error {
	if s.delete != nil {
		return s.delete(id)
	}
	return nil
}
func (s *stubUserRepository) WithTx(_ *gorm.DB) user.Repository {
	return s
}

type stubPermissionChecker struct {
	hasPermission func(roleName, permission string) (bool, error)
}

func (s *stubPermissionChecker) HasPermission(roleName, permission string) (bool, error) {
	if s.hasPermission != nil {
		return s.hasPermission(roleName, permission)
	}
	return false, nil
}

func TestRequirePermission_Forbidden(t *testing.T) {
	userService := &user.Service{UserRepo: &stubUserRepository{}}
	handler := user.Handler{UserService: userService}
	checker := &stubPermissionChecker{
		hasPermission: func(roleName, permission string) (bool, error) {
			assert.Equal(t, "user", roleName)
			assert.Equal(t, "user.read_all", permission)
			return false, nil
		},
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("role", "user")
		c.Next()
	})

	r.GET("/admin/users", middleware.RequirePermission(checker, "user.read_all"), func(c *gin.Context) {
		handler.GetAllUsers(c)
	})

	req, _ := http.NewRequest("GET", "/admin/users", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "error", bodyMap["status"])
	assert.Equal(t, "Forbidden", bodyMap["message"])
}

func TestRequirePermission_Allowed(t *testing.T) {
	checker := &stubPermissionChecker{
		hasPermission: func(roleName, permission string) (bool, error) {
			assert.Equal(t, "admin", roleName)
			assert.Equal(t, "user.read_all", permission)
			return true, nil
		},
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})

	r.GET("/admin/users", middleware.RequirePermission(checker, "user.read_all"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/admin/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestUpdateProfile_Success(t *testing.T) {
	repo := &stubUserRepository{}
	service := &user.Service{UserRepo: repo}
	handler := user.Handler{UserService: service}

	repo.findByID = func(id uint) (*user.User, error) {
		return &user.User{Model: gorm.Model{ID: id}, Name: "Old", Email: "test@mail.com", Role: "user"}, nil
	}
	repo.update = func(u *user.User) error {
		assert.Equal(t, "New Name", u.Name)
		return nil
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPatch, "/auth/profile", strings.NewReader(`{"name":"New Name"}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Set("user_id", uint(1))

	handler.UpdateProfile(c)

	assert.Equal(t, 200, w.Code)
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "success", bodyMap["status"])
	assert.Equal(t, "Profile updated", bodyMap["message"])
	data, ok := bodyMap["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "New Name", data["name"])
}

func TestChangePassword_Success(t *testing.T) {
	repo := &stubUserRepository{}
	refreshRepo := &stubRefreshTokenRepo{}
	service := &user.Service{UserRepo: repo, RefreshTokenRepo: refreshRepo}

	hashed, err := bcrypt.GenerateFromPassword([]byte("secret123"), 14)
	assert.NoError(t, err)

	repo.findByID = func(id uint) (*user.User, error) {
		return &user.User{
			Model:    gorm.Model{ID: id},
			Password: string(hashed),
		}, nil
	}
	repo.update = func(u *user.User) error {
		return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("newsecret123"))
	}
	refreshRepo.deleteByUser = func(userID uint) error {
		assert.Equal(t, uint(1), userID)
		return nil
	}

	err = service.ChangePassword(1, "secret123", "newsecret123")

	assert.NoError(t, err)
}

func TestCreateUser_EmailAlreadyExists(t *testing.T) {
	repo := &stubUserRepository{}
	service := &user.Service{UserRepo: repo}

	repo.findByEmail = func(email string) (*user.User, error) {
		return &user.User{Model: gorm.Model{ID: 10}, Email: email}, nil
	}

	_, err := service.CreateUser(user.CreateUserRequest{
		Name:     "Tester",
		Email:    "test@mail.com",
		Password: "secret123",
	})

	assert.EqualError(t, err, "email already in use")
}

func TestCreateUser_RejectsSuperadminRole(t *testing.T) {
	repo := &stubUserRepository{}
	service := &user.Service{UserRepo: repo}

	_, err := service.CreateUser(user.CreateUserRequest{
		Name:     "Tester",
		Email:    "test@mail.com",
		Password: "secret123",
		Role:     "superadmin",
	})

	assert.EqualError(t, err, "unsupported role")
}

func TestChangePassword_InvalidCurrentPassword(t *testing.T) {
	repo := &stubUserRepository{}
	service := &user.Service{UserRepo: repo, RefreshTokenRepo: &stubRefreshTokenRepo{}}

	hashed, err := bcrypt.GenerateFromPassword([]byte("secret123"), 14)
	assert.NoError(t, err)

	repo.findByID = func(id uint) (*user.User, error) {
		return &user.User{
			Model:    gorm.Model{ID: id},
			Password: string(hashed),
		}, nil
	}

	err = service.ChangePassword(1, "wrong-password", "newsecret123")

	assert.EqualError(t, err, "current password is incorrect")
}

func TestChangePassword_PropagatesRefreshTokenRevocationFailure(t *testing.T) {
	repo := &stubUserRepository{}
	refreshRepo := &stubRefreshTokenRepo{}
	service := &user.Service{UserRepo: repo, RefreshTokenRepo: refreshRepo}

	hashed, err := bcrypt.GenerateFromPassword([]byte("secret123"), 14)
	assert.NoError(t, err)

	repo.findByID = func(id uint) (*user.User, error) {
		return &user.User{
			Model:    gorm.Model{ID: id},
			Password: string(hashed),
		}, nil
	}
	repo.update = func(u *user.User) error { return nil }
	refreshRepo.deleteByUser = func(userID uint) error {
		assert.Equal(t, uint(1), userID)
		return errors.New("delete failed")
	}

	err = service.ChangePassword(1, "secret123", "newsecret123")

	assert.EqualError(t, err, "delete failed")
}

func TestUpdateUser_RoleChangeRevokesRefreshTokens(t *testing.T) {
	repo := &stubUserRepository{}
	refreshRepo := &stubRefreshTokenRepo{}
	service := &user.Service{UserRepo: repo, RefreshTokenRepo: refreshRepo}

	repo.findByID = func(id uint) (*user.User, error) {
		return &user.User{
			Model:              gorm.Model{ID: id},
			Name:               "Old Name",
			Email:              "old@mail.com",
			Role:               "user",
			AccessTokenVersion: 4,
		}, nil
	}
	repo.findByEmail = func(email string) (*user.User, error) {
		return nil, gorm.ErrRecordNotFound
	}
	repo.update = func(u *user.User) error {
		assert.Equal(t, "admin", u.Role)
		assert.Equal(t, uint(5), u.AccessTokenVersion)
		return nil
	}
	refreshRepo.deleteByUser = func(userID uint) error {
		assert.Equal(t, uint(7), userID)
		return nil
	}

	updated, err := service.UpdateUser(7, user.UpdateUserRequest{
		Name:       "New Name",
		Email:      "new@mail.com",
		Role:       "admin",
		IsVerified: true,
	})

	assert.NoError(t, err)
	assert.Equal(t, uint(5), updated.AccessTokenVersion)
}

func TestUpdateUser_RoleChangeReturnsRefreshTokenRevocationError(t *testing.T) {
	repo := &stubUserRepository{}
	refreshRepo := &stubRefreshTokenRepo{}
	service := &user.Service{UserRepo: repo, RefreshTokenRepo: refreshRepo}

	repo.findByID = func(id uint) (*user.User, error) {
		return &user.User{
			Model:              gorm.Model{ID: id},
			Role:               "user",
			AccessTokenVersion: 1,
		}, nil
	}
	repo.findByEmail = func(email string) (*user.User, error) {
		return nil, gorm.ErrRecordNotFound
	}
	repo.update = func(u *user.User) error { return nil }
	refreshRepo.deleteByUser = func(userID uint) error {
		return errors.New("revoke failed")
	}

	_, err := service.UpdateUser(7, user.UpdateUserRequest{
		Name:  "New Name",
		Email: "new@mail.com",
		Role:  "admin",
	})

	assert.EqualError(t, err, "revoke failed")
}

func TestUpdateUser_RejectsSuperadminRole(t *testing.T) {
	repo := &stubUserRepository{}
	service := &user.Service{UserRepo: repo, RefreshTokenRepo: &stubRefreshTokenRepo{}}

	repo.findByID = func(id uint) (*user.User, error) {
		return &user.User{
			Model: gorm.Model{ID: id},
			Role:  "user",
		}, nil
	}

	_, err := service.UpdateUser(7, user.UpdateUserRequest{
		Name:  "New Name",
		Email: "new@mail.com",
		Role:  "superadmin",
	})

	assert.EqualError(t, err, "unsupported role")
}
