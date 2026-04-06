package repositories

import (
	"go-auth-app/config"
	"go-auth-app/models"
)

type RoleRepoDB struct{}

func (r *RoleRepoDB) FindByID(id uint) (*models.Role, error) {
	var role models.Role

	err := config.DB.Preload("Permissions").
		First(&role, id).Error

	if err != nil {
		return nil, err
	}

	return &role, nil
}
