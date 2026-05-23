package user

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(user *User) error
	FindByEmail(email string) (*User, error)
	FindByPhone(phone string) (*User, error)
	FindByID(id uint) (*User, error)
	Update(user *User) error
	UpdateLastLogin(id uint, at time.Time) error
	FindAll() ([]User, error)
	FindAllWithFilter(page, limit int, search, role string) ([]User, int64, error)
	Delete(id uint) error
	WithTx(tx *gorm.DB) Repository
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(user *User) error {
	if strings.TrimSpace(user.PhoneNumber) == "" {
		return r.db.Omit("phone_number").Create(user).Error
	}
	return r.db.Create(user).Error
}

func (r *GormRepository) FindByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (r *GormRepository) FindByPhone(phone string) (*User, error) {
	var user User
	err := r.db.Where("phone_number = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (r *GormRepository) FindByID(id uint) (*User, error) {
	var user User
	err := r.db.Preload("RoleDetails").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (r *GormRepository) Update(user *User) error {
	phoneNumber := interface{}(user.PhoneNumber)
	if strings.TrimSpace(user.PhoneNumber) == "" {
		phoneNumber = nil
	}

	updates := map[string]interface{}{
		"name":                 user.Name,
		"email":                user.Email,
		"phone_number":         phoneNumber,
		"role":                 user.Role,
		"role_id":              user.RoleID,
		"is_verified":          user.IsVerified,
		"phone_verified":       user.PhoneVerified,
		"email_verified":       user.EmailVerified,
		"password_updated_at":  user.PasswordUpdatedAt,
		"last_login_at":        user.LastLoginAt,
		"last_password_change": user.LastPasswordChange,
		"access_token_version": user.AccessTokenVersion,
	}
	if user.Password != "" {
		updates["password"] = user.Password
	}
	return r.db.Model(user).Updates(updates).Error
}

func (r *GormRepository) UpdateLastLogin(id uint, at time.Time) error {
	return r.db.Session(&gorm.Session{SkipDefaultTransaction: true}).
		Model(&User{}).
		Where("id = ?", id).
		Update("last_login_at", at).Error
}

func (r *GormRepository) FindAll() ([]User, error) {
	var users []User
	err := r.db.Preload("RoleDetails").Find(&users).Error
	return users, err
}

func (r *GormRepository) FindAllWithFilter(page, limit int, search, role string) ([]User, int64, error) {
	var users []User
	var total int64

	query := r.db.Model(&User{})

	if search != "" {
		like := "%" + search + "%"
		query = query.Where(userTextSearchCondition(r.db), userTextSearchValues(r.db, like)...)
	}

	if role != "" {
		query = query.Where("role = ?", role)
	}

	query.Count(&total)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	err := query.Preload("RoleDetails").Limit(limit).Offset(offset).Find(&users).Error

	return users, total, err
}

func (r *GormRepository) Delete(id uint) error {
	return r.db.Delete(&User{}, id).Error
}

func (r *GormRepository) WithTx(tx *gorm.DB) Repository {
	return &GormRepository{db: tx}
}

func userTextSearchCondition(db *gorm.DB) string {
	if db.Dialector.Name() == "postgres" {
		return "name ILIKE ? OR email ILIKE ? OR phone_number ILIKE ?"
	}
	return "LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(phone_number) LIKE ?"
}

func userTextSearchValues(db *gorm.DB, like string) []interface{} {
	if db.Dialector.Name() == "postgres" {
		return []interface{}{like, like, like}
	}
	lowerLike := strings.ToLower(like)
	return []interface{}{lowerLike, lowerLike, lowerLike}
}
