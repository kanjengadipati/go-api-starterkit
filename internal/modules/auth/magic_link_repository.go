package auth

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type magicLinkRepository interface {
	Create(token *MagicLinkToken) error
	FindActiveByTokenHash(tokenHash string, now time.Time) (*MagicLinkToken, error)
	Consume(id string, consumedAt time.Time) error
	WithTx(tx *gorm.DB) magicLinkRepository
}

type gormMagicLinkRepository struct {
	db *gorm.DB
}

func newMagicLinkRepository(db *gorm.DB) magicLinkRepository {
	return &gormMagicLinkRepository{db: db}
}

func (r *gormMagicLinkRepository) Create(token *MagicLinkToken) error {
	return r.db.Create(token).Error
}

func (r *gormMagicLinkRepository) FindActiveByTokenHash(tokenHash string, now time.Time) (*MagicLinkToken, error) {
	var token MagicLinkToken
	err := r.db.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token_hash = ? AND consumed_at IS NULL AND expires_at > ?", tokenHash, now).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *gormMagicLinkRepository) Consume(id string, consumedAt time.Time) error {
	return r.db.Model(&MagicLinkToken{}).
		Where("id = ? AND consumed_at IS NULL", id).
		Update("consumed_at", consumedAt).Error
}

func (r *gormMagicLinkRepository) WithTx(tx *gorm.DB) magicLinkRepository {
	return &gormMagicLinkRepository{db: tx}
}
