package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/mail"
	"strings"
	"time"

	"pleco-api/internal/modules/audit"
	userModule "pleco-api/internal/modules/user"
	"pleco-api/internal/otp"
	"pleco-api/internal/services"
	"pleco-api/internal/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	otpPurposeLogin       = "login"
	otpExpiry             = 5 * time.Minute
	otpMaxAttempts        = 5
	legacyDashboardDevice = "nextjs-dashboard"
)

var (
	ErrInvalidOTPChannel = errors.New("unsupported otp channel")
	ErrOTPNotAvailable   = errors.New("unable to send OTP")
	ErrInvalidOTP        = errors.New("invalid or expired OTP")
	ErrOTPRateLimited    = errors.New("too many OTP requests")
	ErrOTPWhatsAppTarget = errors.New("no whatsapp number available for this account")
)

type PasswordlessStartResult struct {
	NextStep string `json:"next_step"`
}

func (s *authService) CheckPasswordlessIdentity(channel, target string) error {
	_, _, err := normalizeOTPTarget(channel, target)
	return err
}

func (s *authService) StartPasswordless(ctx context.Context, channel, target, deviceID, userAgent, ipAddress string) (*PasswordlessStartResult, error) {
	channel, target, err := normalizeOTPTarget(channel, target)
	if err != nil {
		return nil, err
	}

	if user, err := findOTPUser(s.UserRepo, channel, target); err == nil && s.isTrustedDevice(user.ID, userAgent, deviceID) && user.Email != "" {
		token, err := s.generateMagicLinkToken(user.ID, user.Email)
		if err != nil {
			return nil, ErrOTPNotAvailable
		}
		if err := s.EmailSvc.SendMagicLink(user.Email, token); err != nil {
			log.Printf("magic link email failed for %s: %v", user.Email, err)
			return nil, ErrOTPNotAvailable
		}
		s.recordOTPAudit(&user.ID, "MAGIC_LINK_REQUESTED", channel, target, "success", "magic link requested", ipAddress, userAgent)
		return &PasswordlessStartResult{NextStep: "magic_link"}, nil
	}

	if err := s.RequestOTP(ctx, channel, target, ipAddress, userAgent); err != nil {
		return nil, err
	}
	return &PasswordlessStartResult{NextStep: "otp"}, nil
}

func (s *authService) RequestOTP(ctx context.Context, channel, target, ipAddress, userAgent string) error {
	channel, target, err := s.resolveOTPTarget(channel, target)
	if err != nil {
		return err
	}

	if s.OTPRepo == nil {
		return ErrOTPNotAvailable
	}

	delivery, ok := s.OTPChannels[channel]
	if !ok || delivery == nil {
		return ErrInvalidOTPChannel
	}

	cooldown := s.otpTargetCooldown()
	if latest, err := s.OTPRepo.FindLatestActive(channel, target, otpPurposeLogin); err == nil && cooldown > 0 && time.Since(latest.CreatedAt) < cooldown {
		s.recordOTPAudit(nil, "OTP_RATE_LIMITED", channel, target, "failed", "otp request cooldown active", ipAddress, userAgent)
		return ErrOTPRateLimited
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrOTPNotAvailable
	}

	targetRequests := s.otpTargetRequests()
	targetWindow := s.otpTargetWindow()
	count, err := s.OTPRepo.CountRequestsSince(channel, target, time.Now().Add(-targetWindow))
	if err != nil {
		return ErrOTPNotAvailable
	}
	if targetRequests > 0 && count >= int64(targetRequests) {
		s.recordOTPAudit(nil, "OTP_RATE_LIMITED", channel, target, "failed", "otp hourly limit reached", ipAddress, userAgent)
		return ErrOTPRateLimited
	}

	code, err := generateNumericOTP(6)
	if err != nil {
		return ErrOTPNotAvailable
	}

	record := &OTPCode{
		ID:        uuid.NewString(),
		Channel:   channel,
		Target:    target,
		CodeHash:  hashOTP(channel, target, code),
		Purpose:   otpPurposeLogin,
		ExpiresAt: time.Now().Add(otpExpiry),
		Provider:  delivery.ChannelName(),
		CreatedAt: time.Now(),
	}
	if err := s.OTPRepo.Create(record); err != nil {
		return ErrOTPNotAvailable
	}

	if err := delivery.SendOTP(ctx, target, otp.Payload{Code: code, ExpiresIn: otpExpiry}); err != nil {
		log.Printf("otp provider failed channel=%s target=%s provider=%s error=%v", channel, target, delivery.ChannelName(), err)
		if consumeErr := s.OTPRepo.Consume(record.ID); consumeErr != nil {
			log.Printf("otp provider failure cleanup failed channel=%s target=%s otp_id=%s error=%v", channel, target, record.ID, consumeErr)
		}
		s.recordOTPAudit(nil, "OTP_REQUESTED", channel, target, "failed", "otp provider failed", ipAddress, userAgent)
		return ErrOTPNotAvailable
	}

	s.recordOTPAudit(nil, "OTP_REQUESTED", channel, target, "success", "otp requested", ipAddress, userAgent)
	return nil
}

func (s *authService) VerifyMagicLink(tokenString, deviceID, deviceName string, trustedDevice bool, userAgent, ipAddress string) (*AuthTokens, error) {
	if s.MagicLinkRepo == nil {
		return nil, ErrInvalidOTP
	}

	var user *userModule.User
	if err := s.runMagicLinkTx(func(userRepo userModule.Repository, magicRepo magicLinkRepository) error {
		record, err := magicRepo.FindActiveByTokenHash(utils.HashToken(strings.TrimSpace(tokenString)), time.Now())
		if err != nil {
			return err
		}
		foundUser, err := userRepo.FindByID(record.UserID)
		if err != nil {
			return err
		}
		if err := magicRepo.Consume(record.ID, time.Now()); err != nil {
			return err
		}
		user = foundUser
		return nil
	}); err != nil {
		return nil, ErrInvalidOTP
	}

	if trustedDevice {
		if err := s.trustDevice(user.ID, userAgent, deviceID, deviceName); err != nil {
			return nil, ErrInvalidOTP
		}
	}

	tokens, err := s.issueTokens(user.ID, user.Role, user.AccessTokenVersion, deviceID, userAgent, ipAddress)
	if err != nil {
		return nil, err
	}
	_ = s.UserRepo.UpdateLastLogin(user.ID, time.Now())
	s.invalidateUserCache(user.ID)
	s.recordOTPAudit(&user.ID, "LOGIN_SUCCESS", "magic_link", user.Email, "success", "magic link login succeeded", ipAddress, userAgent)
	return tokens, nil
}

func (s *authService) VerifyOTP(ctx context.Context, input VerifyOTPInput) (*AuthTokens, error) {
	channel, target, err := s.resolveOTPTarget(input.Channel, input.Target)
	if err != nil {
		return nil, err
	}
	if s.OTPRepo == nil {
		return nil, ErrInvalidOTP
	}

	record, err := s.OTPRepo.FindLatestActive(channel, target, otpPurposeLogin)
	if err != nil {
		s.recordOTPAudit(nil, "LOGIN_FAILED", channel, target, "failed", "otp not found", input.IPAddress, input.UserAgent)
		return nil, ErrInvalidOTP
	}
	if record.Consumed || time.Now().After(record.ExpiresAt) || record.Attempts >= otpMaxAttempts {
		s.recordOTPAudit(nil, "LOGIN_FAILED", channel, target, "failed", "otp expired or exhausted", input.IPAddress, input.UserAgent)
		return nil, ErrInvalidOTP
	}

	expected := []byte(record.CodeHash)
	actual := []byte(hashOTP(channel, target, strings.TrimSpace(input.OTP)))
	if subtle.ConstantTimeCompare(expected, actual) != 1 {
		_ = s.OTPRepo.IncrementAttempts(record.ID)
		s.recordOTPAudit(nil, "LOGIN_FAILED", channel, target, "failed", "otp verification failed", input.IPAddress, input.UserAgent)
		return nil, ErrInvalidOTP
	}

	var user *userModule.User
	if err := s.runOTPTx(func(userRepo userModule.Repository, otpRepo otpRepository) error {
		var findErr error
		user, findErr = findOTPUser(userRepo, channel, target)
		if findErr != nil {
			if !errors.Is(findErr, gorm.ErrRecordNotFound) {
				return findErr
			}
			created, err := buildOTPUser(channel, target)
			if err != nil {
				return err
			}
			if err := userRepo.Create(created); err != nil {
				return err
			}
			user = created
		}

		markUserVerified(user, channel, target)
		if err := userRepo.Update(user); err != nil {
			return err
		}
		if err := otpRepo.Consume(record.ID); err != nil {
			return err
		}
		if input.TrustedDevice {
			if err := upsertTrustedDevice(otpRepo, user.ID, input.UserAgent, input.DeviceID, input.DeviceName); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, ErrInvalidOTP
	}

	tokens, err := s.issueTokens(user.ID, user.Role, user.AccessTokenVersion, input.DeviceID, input.UserAgent, input.IPAddress)
	if err != nil {
		return nil, err
	}
	if err := s.UserRepo.UpdateLastLogin(user.ID, time.Now()); err != nil {
		return nil, err
	}
	s.invalidateUserCache(user.ID)

	s.recordOTPAudit(&user.ID, "OTP_VERIFIED", channel, target, "success", "otp verified", input.IPAddress, input.UserAgent)
	s.recordOTPAudit(&user.ID, "LOGIN_SUCCESS", channel, target, "success", "otp login succeeded", input.IPAddress, input.UserAgent)
	return tokens, nil
}

type VerifyOTPInput struct {
	Channel       string
	Target        string
	OTP           string
	DeviceID      string
	DeviceName    string
	TrustedDevice bool
	UserAgent     string
	IPAddress     string
}

func (s *authService) runOTPTx(fn func(userRepo userModule.Repository, otpRepo otpRepository) error) error {
	if s.DB == nil {
		return fn(s.UserRepo, s.OTPRepo)
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		return fn(s.UserRepo.WithTx(tx), s.OTPRepo.WithTx(tx))
	})
}

func (s *authService) runMagicLinkTx(fn func(userRepo userModule.Repository, magicRepo magicLinkRepository) error) error {
	if s.DB == nil {
		return fn(s.UserRepo, s.MagicLinkRepo)
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		return fn(s.UserRepo.WithTx(tx), s.MagicLinkRepo.WithTx(tx))
	})
}

func (s *authService) isTrustedDevice(userID uint, userAgent, deviceID string) bool {
	if s.OTPRepo == nil || deviceID == "" {
		return false
	}
	_, err := s.OTPRepo.FindTrustedDevice(userID, hashDevice(userAgent, deviceID))
	if err == nil {
		return true
	}
	if deviceID == legacyDashboardDevice {
		return false
	}
	_, err = s.OTPRepo.FindTrustedDevice(userID, hashDevice(userAgent, legacyDashboardDevice))
	return err == nil
}

func (s *authService) trustDevice(userID uint, userAgent, deviceID, deviceName string) error {
	if s.OTPRepo == nil {
		return nil
	}
	return upsertTrustedDevice(s.OTPRepo, userID, userAgent, deviceID, deviceName)
}

func upsertTrustedDevice(repo otpRepository, userID uint, userAgent, deviceID, deviceName string) error {
	if repo == nil || deviceID == "" {
		return nil
	}
	now := time.Now()
	if deviceName == "" {
		deviceName = deviceID
	}
	return repo.UpsertTrustedDevice(&TrustedDevice{
		ID:         uuid.NewString(),
		UserID:     userID,
		DeviceHash: hashDevice(userAgent, deviceID),
		DeviceName: deviceName,
		LastUsedAt: &now,
		CreatedAt:  now,
	})
}

func normalizeOTPTarget(channel, target string) (string, string, error) {
	channel = strings.ToLower(strings.TrimSpace(channel))
	target = strings.TrimSpace(target)
	switch channel {
	case "email":
		target = strings.ToLower(target)
		if _, err := mail.ParseAddress(target); err != nil {
			return "", "", ErrInvalidOTPChannel
		}
	case "whatsapp":
		if !isE164Phone(target) {
			return "", "", ErrInvalidOTPChannel
		}
	default:
		return "", "", ErrInvalidOTPChannel
	}
	return channel, target, nil
}

func (s *authService) resolveOTPTarget(channel, target string) (string, string, error) {
	channel = strings.ToLower(strings.TrimSpace(channel))
	target = strings.TrimSpace(target)
	if channel != "whatsapp" {
		return normalizeOTPTarget(channel, target)
	}
	if isE164Phone(target) {
		return channel, target, nil
	}

	email := strings.ToLower(target)
	if _, err := mail.ParseAddress(email); err != nil {
		return "", "", ErrInvalidOTPChannel
	}
	if s.UserRepo == nil {
		return "", "", ErrOTPWhatsAppTarget
	}

	user, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", ErrOTPWhatsAppTarget
		}
		return "", "", ErrOTPNotAvailable
	}
	phone := strings.TrimSpace(user.PhoneNumber)
	if !isE164Phone(phone) {
		return "", "", ErrOTPWhatsAppTarget
	}
	return channel, phone, nil
}

func isE164Phone(target string) bool {
	if len(target) < 9 || len(target) > 16 || !strings.HasPrefix(target, "+") {
		return false
	}
	for index, char := range target[1:] {
		if char < '0' || char > '9' {
			return false
		}
		if index == 0 && char == '0' {
			return false
		}
	}
	return true
}

func (s *authService) otpTargetCooldown() time.Duration {
	seconds := s.Cfg.OTPRateLimit.TargetCooldownSeconds
	if seconds == 0 {
		seconds = 60
	}
	return time.Duration(seconds) * time.Second
}

func (s *authService) otpTargetRequests() int {
	requests := s.Cfg.OTPRateLimit.TargetRequests
	if requests == 0 {
		return 5
	}
	return requests
}

func (s *authService) otpTargetWindow() time.Duration {
	seconds := s.Cfg.OTPRateLimit.TargetWindowSeconds
	if seconds == 0 {
		seconds = 3600
	}
	return time.Duration(seconds) * time.Second
}

func generateNumericOTP(length int) (string, error) {
	var builder strings.Builder
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		builder.WriteString(n.String())
	}
	return builder.String(), nil
}

func hashOTP(channel, target, code string) string {
	sum := sha256.Sum256([]byte(channel + ":" + target + ":" + code))
	return hex.EncodeToString(sum[:])
}

func hashDevice(userAgent, deviceID string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(userAgent) + ":" + strings.TrimSpace(deviceID)))
	return hex.EncodeToString(sum[:])
}

func (s *authService) generateMagicLinkToken(userID uint, email string) (string, error) {
	if s.MagicLinkRepo == nil {
		return "", ErrOTPNotAvailable
	}
	token := uuid.NewString()
	now := time.Now()
	if err := s.MagicLinkRepo.Create(&MagicLinkToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		Email:     email,
		TokenHash: utils.HashToken(token),
		ExpiresAt: now.Add(15 * time.Minute),
		CreatedAt: now,
	}); err != nil {
		return "", err
	}
	return token, nil
}

func findOTPUser(repo userModule.Repository, channel, target string) (*userModule.User, error) {
	if channel == "email" {
		return repo.FindByEmail(target)
	}
	return repo.FindByPhone(target)
}

func buildOTPUser(channel, target string) (*userModule.User, error) {
	password, err := services.HashPassword(uuid.NewString())
	if err != nil {
		return nil, err
	}
	user := &userModule.User{
		Name:     displayNameForTarget(target),
		Role:     "user",
		Password: password,
	}
	markUserVerified(user, channel, target)
	return user, nil
}

func markUserVerified(user *userModule.User, channel, target string) {
	if user.Role == "" {
		user.Role = "user"
	}
	user.IsVerified = true
	if channel == "email" {
		user.Email = target
		user.EmailVerified = true
		return
	}
	user.PhoneNumber = target
	user.PhoneVerified = true
	if user.Email == "" {
		user.Email = fmt.Sprintf("%s@phone.pleco.local", strings.TrimPrefix(strings.ReplaceAll(target, "+", ""), "0"))
	}
}

func displayNameForTarget(target string) string {
	if strings.Contains(target, "@") {
		return strings.Split(target, "@")[0]
	}
	return target
}

func (s *authService) recordOTPAudit(userID *uint, action, channel, target, status, description, ipAddress, userAgent string) {
	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: userID,
		Action:      action,
		Resource:    "auth",
		ResourceID:  userID,
		Status:      status,
		Description: fmt.Sprintf("%s channel=%s target=%s", description, channel, target),
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
}
