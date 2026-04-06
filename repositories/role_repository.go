package repositories

import "go-auth-app/models"

type RoleRepository interface {
	FindByID(id uint) (*models.Role, error)
}
