package repositories

import (
	"errors"

	"gorm.io/gorm"

	"go-auth-app/config"
	"go-auth-app/models"
)

type SocialAccountRepositoryImpl struct{}

var _ SocialAccountRepository = (*SocialAccountRepositoryImpl)(nil)

func NewSocialAccountRepository() *SocialAccountRepositoryImpl {
	return &SocialAccountRepositoryImpl{}
}

func (r *SocialAccountRepositoryImpl) Create(socialAccount *models.SocialAccount) error {
	if socialAccount == nil {
		return errors.New("socialAccount cannot be nil")
	}
	return config.DB.Create(socialAccount).Error
}

func (r *SocialAccountRepositoryImpl) FindByProvider(provider, providerUserID string) (*models.SocialAccount, error) {
	if provider == "" || providerUserID == "" {
		return nil, errors.New("provider and providerUserID cannot be empty")
	}
	var account models.SocialAccount
	err := config.DB.Where("provider = ? AND provider_user_id = ?", provider, providerUserID).First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &account, nil
}
