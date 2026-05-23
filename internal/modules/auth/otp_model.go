package auth

import (
	"time"
)

type OTPCode struct {
	ID        string    `gorm:"primaryKey"`
	Channel   string    `gorm:"column:channel"`
	Target    string    `gorm:"column:target"`
	CodeHash  string    `gorm:"column:code_hash"`
	Purpose   string    `gorm:"column:purpose"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	Attempts  int       `gorm:"column:attempts"`
	Consumed  bool      `gorm:"column:consumed"`
	Provider  string    `gorm:"column:provider"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (OTPCode) TableName() string {
	return "otp_codes"
}

type TrustedDevice struct {
	ID         string     `gorm:"primaryKey"`
	UserID     uint       `gorm:"column:user_id"`
	DeviceHash string     `gorm:"column:device_hash"`
	DeviceName string     `gorm:"column:device_name"`
	LastUsedAt *time.Time `gorm:"column:last_used_at"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
}

func (TrustedDevice) TableName() string {
	return "trusted_devices"
}
