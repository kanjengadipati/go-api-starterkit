package user

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name               string
	Email              string `gorm:"unique" json:"email"`
	Password           string `json:"-"`
	Role               string `json:"role"` // user / admin
	RoleID             uint   `json:"role_id"`
	IsVerified         bool   `json:"is_verified"`
	PasswordUpdatedAt  time.Time
	AccessTokenVersion uint `gorm:"default:0" json:"-"`
}
