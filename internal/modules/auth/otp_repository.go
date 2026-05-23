package auth

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type otpRepository interface {
	Create(code *OTPCode) error
	FindLatestActive(channel, target, purpose string) (*OTPCode, error)
	CountRequestsSince(channel, target string, since time.Time) (int64, error)
	IncrementAttempts(id string) error
	Consume(id string) error
	UpsertTrustedDevice(device *TrustedDevice) error
	WithTx(tx *gorm.DB) otpRepository
}

type gormOTPRepository struct {
	db *gorm.DB
}

func newOTPRepository(db *gorm.DB) otpRepository {
	return &gormOTPRepository{db: db}
}

func (r *gormOTPRepository) Create(code *OTPCode) error {
	return r.db.Create(code).Error
}

func (r *gormOTPRepository) FindLatestActive(channel, target, purpose string) (*OTPCode, error) {
	var code OTPCode
	err := r.db.
		Where("channel = ? AND target = ? AND purpose = ? AND consumed = ?", channel, target, purpose, false).
		Order("created_at DESC").
		First(&code).Error
	if err != nil {
		return nil, err
	}
	return &code, nil
}

func (r *gormOTPRepository) CountRequestsSince(channel, target string, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&OTPCode{}).
		Where("channel = ? AND target = ? AND created_at >= ?", channel, target, since).
		Count(&count).Error
	return count, err
}

func (r *gormOTPRepository) IncrementAttempts(id string) error {
	return r.db.Model(&OTPCode{}).Where("id = ?", id).Update("attempts", gorm.Expr("attempts + 1")).Error
}

func (r *gormOTPRepository) Consume(id string) error {
	return r.db.Model(&OTPCode{}).Where("id = ?", id).Updates(map[string]interface{}{"consumed": true}).Error
}

func (r *gormOTPRepository) UpsertTrustedDevice(device *TrustedDevice) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "device_hash"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"device_name":  device.DeviceName,
			"last_used_at": device.LastUsedAt,
		}),
	}).Create(device).Error
}

func (r *gormOTPRepository) WithTx(tx *gorm.DB) otpRepository {
	return &gormOTPRepository{db: tx}
}
