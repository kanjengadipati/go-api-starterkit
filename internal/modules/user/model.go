package user

import (
	"time"

	roleModule "pleco-api/internal/modules/role"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name               string
	Email              string          `gorm:"unique" json:"email"`
	PhoneNumber        string          `gorm:"unique" json:"phone_number,omitempty"`
	Password           string          `json:"-"`
	Role               string          `json:"role"` // user / admin
	RoleID             uint            `json:"role_id"`
	RoleDetails        roleModule.Role `gorm:"foreignKey:RoleID" json:"role_details,omitempty"`
	IsVerified         bool            `json:"is_verified"`
	PhoneVerified      bool            `json:"phone_verified"`
	EmailVerified      bool            `json:"email_verified"`
	PasswordUpdatedAt  time.Time
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
	LastPasswordChange *time.Time `json:"last_password_change_at,omitempty"`
	AccessTokenVersion uint       `gorm:"default:0" json:"-"`
}
