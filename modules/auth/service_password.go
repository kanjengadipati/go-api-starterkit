package auth

import (
	"errors"
	"log"
	"time"

	"go-api-starterkit/modules/audit"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (s *authService) ForgotPassword(email string) error {
	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return errors.New("failed to process password reset request")
	}

	token, err := s.generateResetToken(user.ID, user.Email)
	if err != nil {
		return err
	}

	if err := s.EmailSvc.SendPasswordReset(user.Email, token); err != nil {
		log.Printf("password reset email failed for %s: %v", user.Email, err)
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &user.ID,
		Action:      "forgot_password",
		Resource:    "auth",
		ResourceID:  &user.ID,
		Status:      "success",
		Description: "password reset requested",
	})

	return nil
}

func (s *authService) ResetPassword(tokenString string, newPassword string) error {
	claims, err := s.JWT.ValidateToken(tokenString)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	purpose, ok := claims["purpose"].(string)
	if !ok || purpose != "password_reset" {
		return errors.New("invalid token purpose")
	}

	userIDValue, ok := claims["user_id"].(float64)
	if !ok {
		return errors.New("invalid token")
	}

	issuedAtValue, ok := claims["iat"].(float64)
	if !ok {
		return errors.New("invalid token")
	}

	userID := uint(userIDValue)
	user, err := s.UserRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if user.PasswordUpdatedAt.Unix() > int64(issuedAtValue) {
		return errors.New("token already invalid")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), 14)
	if err != nil {
		return errors.New("failed to update password")
	}

	user.Password = string(hashed)
	user.PasswordUpdatedAt = time.Now()

	if err := s.UserRepo.Update(user); err != nil {
		return err
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &user.ID,
		Action:      "reset_password",
		Resource:    "auth",
		ResourceID:  &user.ID,
		Status:      "success",
		Description: "password reset completed",
	})

	return nil
}

func (s *authService) generateResetToken(userID uint, email string) (string, error) {
	claims := map[string]interface{}{
		"user_id": userID,
		"email":   email,
		"purpose": "password_reset",
	}
	return s.JWT.GenerateCustomClaimsToken(claims, 15*time.Minute)
}
