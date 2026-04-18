package auth

import (
	"context"
	"errors"

	"go-api-starterkit/modules/audit"
	permissionless "go-api-starterkit/modules/social"
	userModule "go-api-starterkit/modules/user"

	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

func (s *authService) SocialLogin(provider string, idToken string, deviceID, userAgent, ipAddress string) (*AuthTokens, error) {
	var email, providerUserID, name, avatar string

	switch provider {
	case "google":
		payload, err := idtoken.Validate(context.Background(), idToken, "")
		if err != nil {
			return nil, errors.New("invalid google token")
		}

		email, _ = payload.Claims["email"].(string)
		name, _ = payload.Claims["name"].(string)
		providerUserID = payload.Subject
		avatar, _ = payload.Claims["picture"].(string)

		emailVerified, ok := payload.Claims["email_verified"].(bool)
		if !ok || !emailVerified {
			return nil, errors.New("email not verified by google")
		}
	default:
		return nil, errors.New("unsupported provider")
	}

	if providerUserID == "" {
		return nil, errors.New("invalid provider id")
	}

	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		user = &userModule.User{
			Email:      email,
			Name:       name,
			Role:       "user",
			IsVerified: true,
		}

		if err := s.UserRepo.Create(user); err != nil {
			return nil, err
		}
	}

	if user.Role == "" {
		user.Role = "user"
		if err := s.UserRepo.Update(user); err != nil {
			return nil, err
		}
	}

	social, err := s.SocialRepo.FindByProvider(provider, providerUserID)
	if err != nil {
		return nil, err
	}

	if social != nil {
		if social.UserID != user.ID {
			return nil, errors.New("social account already linked to another user")
		}
	} else {
		newSocial := &permissionless.SocialAccount{
			UserID:         user.ID,
			Provider:       provider,
			ProviderUserID: providerUserID,
			AvatarURL:      avatar,
		}
		if err := s.SocialRepo.Create(newSocial); err != nil {
			return nil, err
		}
	}

	tokens, err := s.issueTokens(user.ID, user.Role, deviceID, userAgent, ipAddress)
	if err != nil {
		return nil, err
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &user.ID,
		Action:      "social_login",
		Resource:    "auth",
		ResourceID:  &user.ID,
		Status:      "success",
		Description: "social login succeeded",
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})

	return tokens, nil
}
