package token

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"gorm.io/gorm"
)

type EmailVerificationRepository interface {
	Create(token *EmailVerificationToken) error
	FindByToken(token string) (*EmailVerificationToken, error)
	DeleteByID(id uint) error
	DeleteByUserID(userID uint) error
	WithTx(tx *gorm.DB) EmailVerificationRepository
}

type GormEmailVerificationRepository struct {
	db *gorm.DB
}

var _ EmailVerificationRepository = (*GormEmailVerificationRepository)(nil)

func NewEmailVerificationRepository(db *gorm.DB) EmailVerificationRepository {
	return &GormEmailVerificationRepository{db: db}
}

func (r *GormEmailVerificationRepository) Create(token *EmailVerificationToken) error {
	return r.db.Create(token).Error
}

func hashEmailVerificationToken(plain string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(plain)))
	return hex.EncodeToString(sum[:])
}

func (r *GormEmailVerificationRepository) FindByToken(plainToken string) (*EmailVerificationToken, error) {
	var verification EmailVerificationToken
	hash := hashEmailVerificationToken(plainToken)
	if err := r.db.Where("token = ?", hash).First(&verification).Error; err != nil {
		return nil, err
	}

	return &verification, nil
}

func (r *GormEmailVerificationRepository) DeleteByID(id uint) error {
	return r.db.Delete(&EmailVerificationToken{}, id).Error
}

func (r *GormEmailVerificationRepository) DeleteByUserID(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&EmailVerificationToken{}).Error
}

func (r *GormEmailVerificationRepository) WithTx(tx *gorm.DB) EmailVerificationRepository {
	return &GormEmailVerificationRepository{db: tx}
}
