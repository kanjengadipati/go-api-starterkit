package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pleco-api/internal/modules/permission"
	"pleco-api/internal/modules/role"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type stubRoleRepo struct {
	findByID func(id uint) (*role.Role, error)
	findAll  func() ([]role.Role, error)
}

func (s *stubRoleRepo) FindByID(id uint) (*role.Role, error) {
	if s.findByID != nil {
		return s.findByID(id)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubRoleRepo) FindAll() ([]role.Role, error) {
	if s.findAll != nil {
		return s.findAll()
	}
	return nil, nil
}
func (s *stubRoleRepo) WithTx(_ *gorm.DB) role.Repository {
	return s
}

type stubPermissionRepo struct {
	listAll             func() ([]permission.Permission, error)
	listRole            func(roleID uint) ([]string, error)
	allPermissionsExist func(names []string) (bool, error)
	replaceRole         func(roleID uint, permissions []string) error
}

func (s *stubPermissionRepo) HasRolePermission(_, _ string) (bool, error) {
	return false, nil
}

func (s *stubPermissionRepo) ListAllPermissions() ([]permission.Permission, error) {
	if s.listAll != nil {
		return s.listAll()
	}
	return nil, nil
}

func (s *stubPermissionRepo) ListRolePermissions(roleID uint) ([]string, error) {
	if s.listRole != nil {
		return s.listRole(roleID)
	}
	return nil, nil
}

func (s *stubPermissionRepo) AllPermissionsExist(names []string) (bool, error) {
	if s.allPermissionsExist != nil {
		return s.allPermissionsExist(names)
	}
	return true, nil
}

func (s *stubPermissionRepo) ReplaceRolePermissions(roleID uint, permissions []string) error {
	if s.replaceRole != nil {
		return s.replaceRole(roleID, permissions)
	}
	return nil
}
func (s *stubPermissionRepo) WithTx(_ *gorm.DB) permission.Repository {
	return s
}

func TestRoleService_UpdateRolePermissions_NormalizesAndReplaces(t *testing.T) {
	roleRepo := &stubRoleRepo{
		findByID: func(id uint) (*role.Role, error) {
			return &role.Role{ID: id, Name: "admin"}, nil
		},
	}

	var replaced []string
	permissionRepo := &stubPermissionRepo{
		allPermissionsExist: func(names []string) (bool, error) {
			assert.Equal(t, []string{"permission.read", "role.read"}, names)
			return true, nil
		},
		replaceRole: func(roleID uint, permissions []string) error {
			assert.Equal(t, uint(2), roleID)
			replaced = permissions
			return nil
		},
	}

	service := role.NewService(roleRepo, permissionRepo)

	foundRole, permissions, err := service.UpdateRolePermissions(2, []string{"role.read", "role.read", "", "permission.read"})

	assert.NoError(t, err)
	assert.Equal(t, "admin", foundRole.Name)
	assert.Equal(t, []string{"permission.read", "role.read"}, permissions)
	assert.Equal(t, permissions, replaced)
}

func TestRoleService_UpdateRolePermissions_RejectsInvalidPermissions(t *testing.T) {
	roleRepo := &stubRoleRepo{
		findByID: func(id uint) (*role.Role, error) {
			return &role.Role{ID: id, Name: "admin"}, nil
		},
	}
	permissionRepo := &stubPermissionRepo{
		allPermissionsExist: func(names []string) (bool, error) {
			return false, nil
		},
	}

	service := role.NewService(roleRepo, permissionRepo)

	_, _, err := service.UpdateRolePermissions(2, []string{"invalid.permission"})

	assert.EqualError(t, err, "one or more permissions are invalid")
}

func TestRoleHandler_GetRolePermissions_Success(t *testing.T) {
	handler := role.NewHandler(&role.Service{
		RoleRepo: &stubRoleRepo{
			findByID: func(id uint) (*role.Role, error) {
				return &role.Role{ID: id, Name: "admin"}, nil
			},
		},
		PermissionRepo: &stubPermissionRepo{
			listRole: func(roleID uint) ([]string, error) {
				return []string{"role.read", "permission.read"}, nil
			},
		},
	})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/roles/1/permissions", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = req

	handler.GetRolePermissions(c)

	assert.Equal(t, http.StatusOK, w.Code)
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "success", bodyMap["status"])
	assert.Equal(t, "Role permissions fetched", bodyMap["message"])
	data := bodyMap["data"].(map[string]interface{})
	assert.Equal(t, "admin", data["name"])
}

func TestRoleHandler_UpdateRolePermissions_InvalidPermission(t *testing.T) {
	handler := role.NewHandler(role.NewService(
		&stubRoleRepo{
			findByID: func(id uint) (*role.Role, error) {
				return &role.Role{ID: id, Name: "admin"}, nil
			},
		},
		&stubPermissionRepo{
			allPermissionsExist: func(names []string) (bool, error) {
				return false, nil
			},
		},
	))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPut, "/roles/1/permissions", strings.NewReader(`{"permissions":["invalid.permission"]}`))
	req.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = req

	handler.UpdateRolePermissions(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	bodyMap := decodeBodyMap(t, w)
	assert.Equal(t, "error", bodyMap["status"])
	assert.Equal(t, "one or more permissions are invalid", bodyMap["message"])
}

func TestRoleHandler_GetRolePermissions_NotFound(t *testing.T) {
	handler := role.NewHandler(&role.Service{
		RoleRepo: &stubRoleRepo{
			findByID: func(id uint) (*role.Role, error) {
				return nil, gorm.ErrRecordNotFound
			},
		},
		PermissionRepo: &stubPermissionRepo{
			listRole: func(roleID uint) ([]string, error) {
				return nil, errors.New("should not be called")
			},
		},
	})

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/roles/1/permissions", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = req

	handler.GetRolePermissions(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
