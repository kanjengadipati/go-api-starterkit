package auth

import (
	"errors"
	"log"
	"time"

	"pleco-api/internal/modules/audit"
	tokenModule "pleco-api/internal/modules/token"
	"pleco-api/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *authService) ResendVerification(email string) error {
	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return errors.New("failed to process verification request")
	}

	if user.IsVerified {
		return nil
	}

	token := uuid.NewString()
	_ = s.EmailVerificationRepo.DeleteByUserID(user.ID)

	verification := &tokenModule.EmailVerificationToken{
		UserID:    user.ID,
		Token:     utils.HashToken(token),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.EmailVerificationRepo.Create(verification); err != nil {
		return errors.New("failed to process verification request")
	}

	if err := s.EmailSvc.SendVerificationEmail(user.Email, token); err != nil {
		log.Printf("verification resend failed for %s: %v", user.Email, err)
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &user.ID,
		Action:      "resend_verification",
		Resource:    "auth",
		ResourceID:  &user.ID,
		Status:      "success",
		Description: "verification email resent",
	})

	return nil
}

func (s *authService) VerifyEmail(token string) error {
	verification, err := s.EmailVerificationRepo.FindByToken(token)
	if err != nil {
		return errors.New("invalid token")
	}

	if time.Now().After(verification.ExpiresAt) {
		return errors.New("token expired")
	}

	user, err := s.UserRepo.FindByID(verification.UserID)
	if err != nil {
		return err
	}

	if user.IsVerified {
		_ = s.EmailVerificationRepo.DeleteByID(verification.ID)
		return nil
	}

	user.IsVerified = true
	if err := s.UserRepo.Update(user); err != nil {
		return err
	}

	_ = s.EmailVerificationRepo.DeleteByID(verification.ID)

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &user.ID,
		Action:      "verify_email",
		Resource:    "auth",
		ResourceID:  &user.ID,
		Status:      "success",
		Description: "email verified",
	})

	return nil
}
