package models

type Role struct {
	ID          uint
	Name        string
	Permissions []Permission `gorm:"many2many:role_permissions"`
}
