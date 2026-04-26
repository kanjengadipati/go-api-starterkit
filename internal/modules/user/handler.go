package user

import (
	"strconv"

	"pleco-api/internal/httpx"
	"pleco-api/internal/modules/audit"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	UserService *Service
}

func NewHandler(userService *Service) *Handler {
	return &Handler{UserService: userService}
}

func (h *Handler) GetAllUsers(c *gin.Context) {
	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}

	search := c.Query("search")
	role := c.Query("role")

	users, total, err := h.UserService.GetAllUsers(page, limit, search, role)
	if err != nil {
		httpx.Error(c, 500, "Failed to fetch users")
		return
	}

	httpx.Success(c, 200, "Users fetched", ToUserResponseList(users), gin.H{
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

func (h *Handler) GetUserByID(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Error(c, 400, "Invalid user id")
		return
	}

	user, err := h.UserService.GetUserByID(uint(id64))
	if err != nil {
		httpx.Error(c, 404, "User not found")
		return
	}

	httpx.Success(c, 200, "User fetched", ToUserResponse(*user), nil)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var input CreateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	user, err := h.UserService.CreateUser(input)
	if err != nil {
		httpx.Error(c, 400, err.Error())
		return
	}

	h.logAudit(c, "create_user", "user", &user.ID, "success", "admin created user")

	httpx.Success(c, 201, "User created", ToUserResponse(*user), nil)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httpx.Error(c, 400, "Invalid user id")
		return
	}

	var input UpdateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	user, err := h.UserService.UpdateUser(uint(id64), input)
	if err != nil {
		httpx.Error(c, 400, err.Error())
		return
	}

	h.logAudit(c, "update_user", "user", &user.ID, "success", "admin updated user")

	httpx.Success(c, 200, "User updated", ToUserResponse(*user), nil)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	userID, ok := userIDValue.(uint)
	if !exists || !ok {
		httpx.Error(c, 401, "Unauthorized")
		return
	}

	var input UpdateProfileRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	user, err := h.UserService.UpdateProfile(userID, input)
	if err != nil {
		httpx.Error(c, 400, err.Error())
		return
	}

	h.logAudit(c, "update_profile", "user", &user.ID, "success", "user updated own profile")

	httpx.Success(c, 200, "Profile updated", ToUserResponse(*user), nil)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	userID, ok := userIDValue.(uint)
	if !exists || !ok {
		httpx.Error(c, 401, "Unauthorized")
		return
	}

	var input ChangePasswordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	if err := h.UserService.ChangePassword(userID, input.CurrentPassword, input.NewPassword); err != nil {
		httpx.Error(c, 400, err.Error())
		return
	}

	h.logAudit(c, "change_password", "user", &userID, "success", "user changed password")

	httpx.Success(c, 200, "Password updated", nil, nil)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")

	id64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		httpx.Error(c, 400, "Invalid user id")
		return
	}

	if err := h.UserService.DeleteUser(uint(id64)); err != nil {
		httpx.Error(c, 500, "Failed")
		return
	}

	targetID := uint(id64)
	h.logAudit(c, "delete_user", "user", &targetID, "success", "admin deleted user")

	httpx.Success(c, 200, "Deleted", nil, nil)
}

func (h *Handler) logAudit(c *gin.Context, action, resource string, resourceID *uint, status, description string) {
	if h.UserService == nil || h.UserService.AuditSvc == nil {
		return
	}

	var actorUserID *uint
	if value, exists := c.Get("user_id"); exists {
		if parsed, ok := value.(uint); ok {
			actorUserID = &parsed
		}
	}

	h.UserService.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: actorUserID,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Status:      status,
		Description: description,
		IPAddress:   c.ClientIP(),
		UserAgent:   c.GetHeader("User-Agent"),
	})
}
