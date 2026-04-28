package auth

import (
	"net/http"
	"sync"
	"time"

	"pleco-api/internal/config"
	"pleco-api/internal/modules/audit"
	permissionless "pleco-api/internal/modules/social"
	tokenModule "pleco-api/internal/modules/token"
	userModule "pleco-api/internal/modules/user"
	"pleco-api/internal/services"

	"gorm.io/gorm"
)

type userRepositoryTx interface {
	FindByID(id uint) (*userModule.User, error)
	Update(user *userModule.User) error
	WithTx(tx *gorm.DB) userModule.Repository
}

type refreshTokenRepositoryTx interface {
	Save(token *tokenModule.RefreshToken) error
	DeleteByUser(userID uint) error
	DeleteByUserAndDevice(userID uint, deviceID string) error
	DeleteByUserExceptDevice(userID uint, deviceID string) error
	WithTx(tx *gorm.DB) tokenModule.RefreshTokenRepository
}

type emailVerificationRepositoryTx interface {
	Create(token *tokenModule.EmailVerificationToken) error
	DeleteByID(id uint) error
	DeleteByUserID(userID uint) error
	WithTx(tx *gorm.DB) tokenModule.EmailVerificationRepository
}

type socialRepositoryTx interface {
	Create(socialAccount *permissionless.SocialAccount) error
	FindByProvider(provider string, providerID string) (*permissionless.SocialAccount, error)
	WithTx(tx *gorm.DB) permissionless.Repository
}

type AuthService interface {
	Register(user *userModule.User, password string) error
	Login(email, password, deviceID, userAgent, ipAddress string) (*AuthTokens, error)
	Logout(userID uint, deviceID string) error
	LogoutAll(userID uint, userAgent, ipAddress string) error
	LogoutOtherSessions(userID uint, currentDeviceID, userAgent, ipAddress string) (*AuthTokens, error)
	RefreshToken(oldRefreshToken string) (*AuthTokens, error)
	GetProfile(userID uint) (*userModule.User, error)
	ListSessions(userID uint, currentDeviceID string) ([]Session, error)
	RevokeSession(userID, sessionID uint, userAgent, ipAddress string) error
	ResendVerification(email string) error
	VerifyEmail(token string) error
	ForgotPassword(email string) error
	ResetPassword(token string, newPassword string) error
	SocialLogin(provider string, idToken string, deviceID, userAgent, ipAddress string) (*AuthTokens, error)
}

type authService struct {
	DB                    *gorm.DB
	UserRepo              userModule.Repository
	RefreshTokenRepo      tokenModule.RefreshTokenRepository
	EmailVerificationRepo tokenModule.EmailVerificationRepository
	SocialRepo            permissionless.Repository
	JWT                   *services.JWTService
	EmailSvc              services.EmailService
	AuditSvc              *audit.Service
	SocialCfg             config.SocialConfig
	socialHTTPClient      *http.Client
	appleKeysCache        *appleJWKSet
	appleKeysCacheTime    time.Time
	appleKeysMutex        sync.RWMutex
}

// AuthServiceOption configures optional authService dependencies (e.g. tests).
type AuthServiceOption func(*authService)

// WithSocialHTTPClient overrides the HTTP client used for social provider requests.
func WithSocialHTTPClient(c *http.Client) AuthServiceOption {
	return func(s *authService) {
		s.socialHTTPClient = c
	}
}

var _ AuthService = (*authService)(nil)

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Session struct {
	ID        uint      `json:"id"`
	DeviceID  string    `json:"device_id"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiredAt time.Time `json:"expired_at"`
	IsCurrent bool      `json:"is_current"`
}

const (
	TokenAccess  = "access"
	TokenRefresh = "refresh"
)

func NewService(db *gorm.DB, cfg config.AppConfig, _ *userModule.Service, jwtService *services.JWTService, auditSvc *audit.Service) AuthService {
	userRepo := userModule.NewRepository(db)
	refreshTokenRepo := tokenModule.NewRefreshTokenRepository(db)
	emailVerificationRepo := tokenModule.NewEmailVerificationRepository(db)
	socialRepo := permissionless.NewRepository(db)
	emailSvc := services.NewEmailService(cfg.Email)

	return NewAuthService(
		db,
		userRepo,
		refreshTokenRepo,
		emailVerificationRepo,
		socialRepo,
		jwtService,
		emailSvc,
		auditSvc,
		cfg.Social,
	)
}

func NewAuthService(
	db *gorm.DB,
	userRepo userModule.Repository,
	refreshRepo tokenModule.RefreshTokenRepository,
	emailRepo tokenModule.EmailVerificationRepository,
	socialRepo permissionless.Repository,
	jwt *services.JWTService,
	emailSvc services.EmailService,
	auditSvc *audit.Service,
	socialCfg config.SocialConfig,
	opts ...AuthServiceOption,
) AuthService {
	s := &authService{
		DB:                    db,
		UserRepo:              userRepo,
		RefreshTokenRepo:      refreshRepo,
		EmailVerificationRepo: emailRepo,
		SocialRepo:            socialRepo,
		JWT:                   jwt,
		EmailSvc:              emailSvc,
		AuditSvc:              auditSvc,
		SocialCfg:             socialCfg,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
	if s.socialHTTPClient == nil {
		s.socialHTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	return s
}

func (s *authService) runUserRefreshTx(fn func(userRepo userRepositoryTx, refreshRepo refreshTokenRepositoryTx) error) error {
	if s.DB == nil {
		return fn(s.UserRepo, s.RefreshTokenRepo)
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		return fn(s.UserRepo.WithTx(tx), s.RefreshTokenRepo.WithTx(tx))
	})
}

func (s *authService) runVerificationTx(fn func(userRepo userModule.Repository, emailRepo emailVerificationRepositoryTx) error) error {
	if s.DB == nil {
		return fn(s.UserRepo, s.EmailVerificationRepo)
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		return fn(s.UserRepo.WithTx(tx), s.EmailVerificationRepo.WithTx(tx))
	})
}

func (s *authService) runUserSocialTx(fn func(userRepo userModule.Repository, socialRepo socialRepositoryTx) error) error {
	if s.DB == nil {
		return fn(s.UserRepo, s.SocialRepo)
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		return fn(s.UserRepo.WithTx(tx), s.SocialRepo.WithTx(tx))
	})
}
