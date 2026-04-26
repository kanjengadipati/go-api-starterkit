package role

import (
	"errors"
	"strconv"

	"pleco-api/internal/httpx"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	RoleService *Service
}

func NewHandler(roleService *Service) *Handler {
	return &Handler{RoleService: roleService}
}

func (h *Handler) GetRoles(c *gin.Context) {
	roles, err := h.RoleService.ListRoles()
	if err != nil {
		httpx.Error(c, 500, "Failed to fetch roles")
		return
	}

	httpx.Success(c, 200, "Roles fetched", roles, nil)
}

func (h *Handler) GetPermissions(c *gin.Context) {
	permissions, err := h.RoleService.ListPermissions()
	if err != nil {
		httpx.Error(c, 500, "Failed to fetch permissions")
		return
	}

	httpx.Success(c, 200, "Permissions fetched", permissions, nil)
}

func (h *Handler) GetRolePermissions(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Error(c, 400, "Invalid role id")
		return
	}

	role, permissions, err := h.RoleService.GetRolePermissions(uint(roleID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpx.Error(c, 404, "Role not found")
			return
		}
		httpx.Error(c, 500, "Failed to fetch role permissions")
		return
	}

	httpx.Success(c, 200, "Role permissions fetched", RolePermissionsResponse{
		ID:          role.ID,
		Name:        role.Name,
		Permissions: permissions,
	}, nil)
}

func (h *Handler) UpdateRolePermissions(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Error(c, 400, "Invalid role id")
		return
	}

	var input UpdateRolePermissionsRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	role, permissions, err := h.RoleService.UpdateRolePermissions(uint(roleID), input.Permissions)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpx.Error(c, 404, "Role not found")
			return
		}
		if err.Error() == "one or more permissions are invalid" {
			httpx.Error(c, 400, err.Error())
			return
		}
		httpx.Error(c, 500, "Failed to update role permissions")
		return
	}

	httpx.Success(c, 200, "Role permissions updated", RolePermissionsResponse{
		ID:          role.ID,
		Name:        role.Name,
		Permissions: permissions,
	}, nil)
}
