package user

import "time"

type UserResponse struct {
	ID                 uint                 `json:"id"`
	Name               string               `json:"name"`
	Email              string               `json:"email"`
	PhoneNumber        string               `json:"phone_number,omitempty"`
	Role               string               `json:"role"`
	RoleID             uint                 `json:"role_id"`
	RoleDetails        *RoleDetailsResponse `json:"role_details,omitempty"`
	IsVerified         bool                 `json:"is_verified"`
	PhoneVerified      bool                 `json:"phone_verified"`
	EmailVerified      bool                 `json:"email_verified"`
	LastLoginAt        *time.Time           `json:"last_login_at,omitempty"`
	LastPasswordChange *time.Time           `json:"last_password_change_at,omitempty"`
}

type RoleDetailsResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type UserPermissionsResponse struct {
	ID          uint     `json:"id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type CreateUserRequest struct {
	Name       string `json:"name" binding:"required,min=3"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	Role       string `json:"role" binding:"omitempty,oneof=admin user"`
	IsVerified bool   `json:"is_verified"`
}

type UpdateUserRequest struct {
	Name       string `json:"name" binding:"required,min=3"`
	Email      string `json:"email" binding:"required,email"`
	Role       string `json:"role" binding:"required,oneof=admin user"`
	IsVerified bool   `json:"is_verified"`
}

type UpdateProfileRequest struct {
	Name        string `json:"name" binding:"required,min=3"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,e164"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=8"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

func ToUserResponse(user User) UserResponse {
	response := UserResponse{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		PhoneNumber:   user.PhoneNumber,
		Role:          user.Role,
		RoleID:        user.RoleID,
		IsVerified:    user.IsVerified,
		PhoneVerified: user.PhoneVerified,
		EmailVerified: user.EmailVerified,
	}
	response.LastLoginAt = user.LastLoginAt
	response.LastPasswordChange = user.LastPasswordChange
	if user.RoleDetails.ID != 0 {
		response.RoleDetails = &RoleDetailsResponse{
			ID:   user.RoleDetails.ID,
			Name: user.RoleDetails.Name,
		}
	}
	return response
}

func ToUserResponseList(users []User) []UserResponse {
	result := make([]UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, ToUserResponse(user))
	}

	return result
}
