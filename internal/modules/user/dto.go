package user

type UserResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	IsVerified bool   `json:"is_verified"`
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
	Name string `json:"name" binding:"required,min=3"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=8"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

func ToUserResponse(user User) UserResponse {
	return UserResponse{
		ID:         user.ID,
		Name:       user.Name,
		Email:      user.Email,
		Role:       user.Role,
		IsVerified: user.IsVerified,
	}
}

func ToUserResponseList(users []User) []UserResponse {
	result := make([]UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, ToUserResponse(user))
	}

	return result
}
