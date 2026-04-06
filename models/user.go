package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID                uint
	Name              string
	Email             string `gorm:"unique"`
	Password          string `json:"-"`
	Role              string // user / admin
	RoleID            uint
	RefreshToken      string
	IsVerified        bool
	PasswordUpdatedAt time.Time
}
