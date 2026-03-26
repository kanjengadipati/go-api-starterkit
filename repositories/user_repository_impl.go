package repositories

import (
	"go-auth-app/config"
	"go-auth-app/models"
)

type UserRepoDB struct{}

func (r *UserRepoDB) Create(user *models.User) error {
	return config.DB.Create(user).Error
}

func (r *UserRepoDB) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := config.DB.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *UserRepoDB) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := config.DB.First(&user, id).Error
	return &user, err
}

func (r *UserRepoDB) Update(user *models.User) error {
	return config.DB.Save(user).Error
}

func (r *UserRepoDB) FindAll() ([]models.User, error) {
	var users []models.User
	err := config.DB.Find(&users).Error
	return users, err
}

func (r *UserRepoDB) Delete(id uint) error {
	return config.DB.Delete(&models.User{}, id).Error
}

func (r *UserRepoDB) FindAllWithFilter(page, limit int, search, role string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := config.DB.Model(&models.User{})

	// 🔍 search (name/email)
	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 🎯 filter role
	if role != "" {
		// apply role filter, but always exclude admin
		query = query.Where("role = ? AND role != ?", role, "admin")
	} else {
		// if no specific role filter, still exclude admin from results
		query = query.Where("role != ?", "admin")
	}

	// count total
	query.Count(&total)

	// max limit
	if limit > 100 {
		limit = 100
	}

	// pagination
	offset := (page - 1) * limit

	err := query.Limit(limit).Offset(offset).Find(&users).Error

	return users, total, err
}
