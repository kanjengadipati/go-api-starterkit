package auth

import (
	"errors"
	"time"

	"pleco-api/internal/modules/audit"
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

	if err := s.RefreshTokenRepo.DeleteByID(token.ID); err != nil {
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
	if err := s.RefreshTokenRepo.DeleteByUser(userID); err != nil {
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

func (s *authService) LogoutOtherSessions(userID uint, currentDeviceID, userAgent, ipAddress string) error {
	if err := s.RefreshTokenRepo.DeleteByUserExceptDevice(userID, currentDeviceID); err != nil {
		return err
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

	return nil
}

func (s *authService) ListSessions(userID uint, currentDeviceID string) ([]Session, error) {
	tokens, err := s.RefreshTokenRepo.FindByUser(userID)
	if err != nil {
		return nil, err
	}

	sessions := make([]Session, 0, len(tokens))
	for _, token := range tokens {
		sessions = append(sessions, Session{
			ID:        token.ID,
			DeviceID:  token.DeviceID,
			UserAgent: token.UserAgent,
			IPAddress: token.IPAddress,
			CreatedAt: token.CreatedAt,
			UpdatedAt: token.UpdatedAt,
			ExpiredAt: token.ExpiredAt,
			IsCurrent: currentDeviceID != "" && token.DeviceID == currentDeviceID,
		})
	}

	return sessions, nil
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

	if time.Now().After(matchedToken.ExpiredAt) {
		return nil, ErrRefreshTokenExpired
	}

	if err := s.RefreshTokenRepo.DeleteByID(matchedToken.ID); err != nil {
		return nil, err
	}

	user, err := s.UserRepo.FindByID(uid)
	if err != nil {
		return nil, err
	}

	newTokens, err := s.issueTokens(uid, user.Role, user.AccessTokenVersion, matchedToken.DeviceID, matchedToken.UserAgent, matchedToken.IPAddress)
	if err != nil {
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
