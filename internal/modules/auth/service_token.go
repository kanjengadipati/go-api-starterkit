package auth

import (
	"errors"
	"time"

	"pleco-api/internal/modules/audit"
	token "pleco-api/internal/modules/token"
	"pleco-api/internal/utils"

	"gorm.io/gorm"
)

func (s *authService) Logout(userID uint, deviceID string) error {
	token, err := s.RefreshTokenRepo.FindByUserAndDevice(userID, deviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.AuditSvc.SafeRecord(audit.RecordInput{
				ActorUserID: &userID,
				Action:      "logout",
				Resource:    "auth",
				ResourceID:  &userID,
				Status:      "success",
				Description: "logout requested without stored refresh token",
			})
			return nil
		}
		return err
	}

	if err := s.RefreshTokenRepo.DeleteByUserAndDevice(userID, deviceID); err != nil {
		return err
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &userID,
		Action:      "logout",
		Resource:    "auth",
		ResourceID:  &userID,
		Status:      "success",
		Description: "user logged out",
		UserAgent:   token.UserAgent,
		IPAddress:   token.IPAddress,
	})

	return nil
}

func (s *authService) LogoutAll(userID uint, userAgent, ipAddress string) error {
	if err := s.runUserRefreshTx(func(userRepo userRepositoryTx, refreshRepo refreshTokenRepositoryTx) error {
		user, err := userRepo.FindByID(userID)
		if err != nil {
			return err
		}

		user.AccessTokenVersion++
		if err := userRepo.Update(user); err != nil {
			return err
		}

		return refreshRepo.DeleteByUser(userID)
	}); err != nil {
		return err
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &userID,
		Action:      "logout_all",
		Resource:    "auth",
		ResourceID:  &userID,
		Status:      "success",
		Description: "all sessions revoked",
		UserAgent:   userAgent,
		IPAddress:   ipAddress,
	})

	return nil
}

func (s *authService) LogoutOtherSessions(userID uint, currentDeviceID, userAgent, ipAddress string) (*AuthTokens, error) {
	user, err := s.UserRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	user.AccessTokenVersion++

	if err := s.runUserRefreshTx(func(userRepo userRepositoryTx, refreshRepo refreshTokenRepositoryTx) error {
		if err := userRepo.Update(user); err != nil {
			return err
		}
		return refreshRepo.DeleteByUserExceptDevice(userID, currentDeviceID)
	}); err != nil {
		return nil, err
	}

	tokens, err := s.issueTokens(user.ID, user.Role, user.AccessTokenVersion, currentDeviceID, userAgent, ipAddress)
	if err != nil {
		return nil, err
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &userID,
		Action:      "logout_other_sessions",
		Resource:    "auth",
		ResourceID:  &userID,
		Status:      "success",
		Description: "other sessions revoked",
		UserAgent:   userAgent,
		IPAddress:   ipAddress,
	})

	return tokens, nil
}

func (s *authService) ListSessions(userID uint, currentDeviceID string) ([]Session, error) {
	tokens, err := s.RefreshTokenRepo.FindByUser(userID)
	if err != nil {
		return nil, err
	}

	trustedByHash := map[string]TrustedDevice{}
	if s.OTPRepo != nil {
		trustedDevices, err := s.OTPRepo.FindTrustedDevicesByUser(userID)
		if err != nil {
			return nil, err
		}
		for _, device := range trustedDevices {
			trustedByHash[device.DeviceHash] = device
		}
	}

	sessions := make([]Session, 0, len(tokens))
	for _, token := range tokens {
		trusted, ok := trustedByHash[hashDevice(token.UserAgent, token.DeviceID)]
		sessions = append(sessions, Session{
			ID:              token.ID,
			DeviceID:        token.DeviceID,
			DeviceName:      trusted.DeviceName,
			TrustedDeviceID: trusted.ID,
			UserAgent:       token.UserAgent,
			IPAddress:       token.IPAddress,
			CreatedAt:       token.CreatedAt,
			UpdatedAt:       token.UpdatedAt,
			ExpiredAt:       token.ExpiredAt,
			IsCurrent:       currentDeviceID != "" && token.DeviceID == currentDeviceID,
			IsTrusted:       ok,
			TrustedAt:       trustedPointer(ok, trusted.CreatedAt),
			LastTrustedAt:   trusted.LastUsedAt,
		})
	}

	return sessions, nil
}

func trustedPointer(ok bool, at time.Time) *time.Time {
	if !ok {
		return nil
	}
	return &at
}

func (s *authService) RevokeSession(userID, sessionID uint, userAgent, ipAddress string) error {
	session, err := s.RefreshTokenRepo.FindByID(sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionNotFound
		}
		return err
	}
	if session.UserID != userID {
		return ErrSessionNotFound
	}

	if err := s.RefreshTokenRepo.DeleteByUserAndID(userID, sessionID); err != nil {
		return err
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &userID,
		Action:      "revoke_session",
		Resource:    "auth",
		ResourceID:  &userID,
		Status:      "success",
		Description: "session revoked",
		UserAgent:   userAgent,
		IPAddress:   ipAddress,
	})

	return nil
}

func (s *authService) RevokeTrustedDevice(userID uint, trustedDeviceID string, userAgent, ipAddress string) error {
	if s.OTPRepo == nil || trustedDeviceID == "" {
		return ErrSessionNotFound
	}
	if err := s.OTPRepo.DeleteTrustedDevice(userID, trustedDeviceID); err != nil {
		return err
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &userID,
		Action:      "revoke_trusted_device",
		Resource:    "auth",
		ResourceID:  &userID,
		Status:      "success",
		Description: "trusted device removed",
		UserAgent:   userAgent,
		IPAddress:   ipAddress,
	})

	return nil
}

func (s *authService) RefreshToken(oldRefreshToken string) (*AuthTokens, error) {
	claims, err := s.JWT.ValidateToken(oldRefreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if claims["type"] != TokenRefresh {
		return nil, ErrInvalidTokenType
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, ErrInvalidTokenClaims
	}
	uid := uint(userID)

	oldHash := utils.HashToken(oldRefreshToken)
	matchedToken, err := s.RefreshTokenRepo.FindByTokenHash(oldHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}
	if matchedToken.UserID != uid {
		return nil, ErrInvalidRefreshToken
	}
	if matchedToken.RevokedAt != nil {
		if err := s.handleRefreshTokenReuse(matchedToken); err != nil {
			return nil, err
		}
		return nil, ErrRefreshTokenReuse
	}

	if time.Now().After(matchedToken.ExpiredAt) {
		return nil, ErrRefreshTokenExpired
	}

	var newTokens *AuthTokens
	if err := s.runUserRefreshTx(func(userRepo userRepositoryTx, refreshRepo refreshTokenRepositoryTx) error {
		currentUser, err := userRepo.FindByID(uid)
		if err != nil {
			return err
		}

		issued, replacementToken, err := s.buildTokenPair(
			uid,
			currentUser.Role,
			currentUser.AccessTokenVersion,
			matchedToken.DeviceID,
			matchedToken.UserAgent,
			matchedToken.IPAddress,
			matchedToken.FamilyID,
			&matchedToken.ID,
		)
		if err != nil {
			return err
		}

		if err := s.persistRefreshToken(refreshRepo, uid, matchedToken.DeviceID, false, replacementToken); err != nil {
			return err
		}

		if err := refreshRepo.RevokeByID(matchedToken.ID, &replacementToken.ID, "rotated"); err != nil {
			return err
		}

		newTokens = issued
		return nil
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if reuseErr := s.handleRefreshTokenReuse(matchedToken); reuseErr != nil {
				return nil, reuseErr
			}
			return nil, ErrRefreshTokenReuse
		}
		return nil, err
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &uid,
		Action:      "refresh_token",
		Resource:    "auth",
		ResourceID:  &uid,
		Status:      "success",
		Description: "refresh token rotated",
		UserAgent:   matchedToken.UserAgent,
		IPAddress:   matchedToken.IPAddress,
	})

	return newTokens, nil
}

func (s *authService) handleRefreshTokenReuse(matchedToken *token.RefreshToken) error {
	if matchedToken == nil {
		return nil
	}

	return s.runUserRefreshTx(func(userRepo userRepositoryTx, refreshRepo refreshTokenRepositoryTx) error {
		user, err := userRepo.FindByID(matchedToken.UserID)
		if err != nil {
			return err
		}

		user.AccessTokenVersion++
		if err := userRepo.Update(user); err != nil {
			return err
		}

		if matchedToken.FamilyID == "" {
			return refreshRepo.DeleteByUser(matchedToken.UserID)
		}
		return refreshRepo.RevokeFamily(matchedToken.UserID, matchedToken.FamilyID, "reuse_detected")
	})
}
