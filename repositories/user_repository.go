package repositories

import "go-auth-app/models"

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	Update(user *models.User) error
	FindAll() ([]models.User, error)
	FindAllWithFilter(page, limit int, search, role string) ([]models.User, int64, error)
	Delete(id uint) error
}
