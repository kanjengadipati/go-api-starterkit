package models

type SocialAccount struct {
	ID             uint
	UserID         uint
	Provider       string // google, github
	ProviderUserID string
	AvatarURL      string
}
