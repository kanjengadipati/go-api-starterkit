package role

import (
	"errors"
	"strconv"

	"pleco-api/internal/domain"
	"pleco-api/internal/httpx"
	"pleco-api/internal/modules/audit"

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

func (h *Handler) GetRoleByID(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Error(c, 400, "Invalid role id")
		return
	}

	role, err := h.RoleService.FindByID(uint(roleID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpx.Error(c, 404, "Role not found")
			return
		}
		httpx.Error(c, 500, "Failed to fetch role")
		return
	}

	httpx.Success(c, 200, "Role fetched", role, nil)
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
		if errors.Is(err, domain.ErrInvalidPermission) {
			httpx.HandleError(c, err)
			return
		}
		httpx.Error(c, 500, "Failed to update role permissions")
		return
	}

	if h.RoleService.AuditSvc != nil {
		var actorUserID *uint
		if value, exists := c.Get("user_id"); exists {
			if parsed, ok := value.(uint); ok {
				actorUserID = &parsed
			}
		}

		h.RoleService.AuditSvc.SafeRecord(audit.RecordInput{
			ActorUserID: actorUserID,
			Action:      "update_role_permissions",
			Resource:    "role",
			ResourceID:  &role.ID,
			Status:      "success",
			Description: "admin updated role permissions for " + role.Name,
			IPAddress:   c.ClientIP(),
			UserAgent:   c.GetHeader("User-Agent"),
		})
	}

	httpx.Success(c, 200, "Role permissions updated", RolePermissionsResponse{
		ID:          role.ID,
		Name:        role.Name,
		Permissions: permissions,
	}, nil)
}
