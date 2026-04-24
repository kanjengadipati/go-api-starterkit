package auth

import (
	"go-api-starterkit/internal/config"
	"go-api-starterkit/internal/modules/audit"
	"go-api-starterkit/internal/modules/permission"
	permissionless "go-api-starterkit/internal/modules/social"
	tokenModule "go-api-starterkit/internal/modules/token"
	userModule "go-api-starterkit/internal/modules/user"
	"go-api-starterkit/internal/services"
	"time"

	"gorm.io/gorm"
)

type AuthService interface {
	Register(user *userModule.User, password string) error
	Login(email, password, deviceID, userAgent, ipAddress string) (*AuthTokens, error)
	Logout(userID uint, deviceID string) error
	LogoutAll(userID uint, userAgent, ipAddress string) error
	LogoutOtherSessions(userID uint, currentDeviceID, userAgent, ipAddress string) error
	RefreshToken(oldRefreshToken string) (*AuthTokens, error)
	GetProfile(userID uint) (*userModule.User, error)
	ListSessions(userID uint, currentDeviceID string) ([]Session, error)
	RevokeSession(userID, sessionID uint, userAgent, ipAddress string) error
	ResendVerification(email string) error
	VerifyEmail(token string) error
	ForgotPassword(email string) error
	ResetPassword(token string, newPassword string) error
	SocialLogin(provider string, idToken string, deviceID, userAgent, ipAddress string) (*AuthTokens, error)
	GetPermissions(roleID uint) ([]string, error)
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
	PermissionSvc         *permission.Service
	SocialCfg             config.SocialConfig
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

func NewService(db *gorm.DB, cfg config.AppConfig, _ *userModule.Service, jwtService *services.JWTService, auditSvc *audit.Service, permissionSvc *permission.Service) AuthService {
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
		permissionSvc,
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
	permissionSvc *permission.Service,
	socialCfg config.SocialConfig,
) AuthService {
	return &authService{
		DB:                    db,
		UserRepo:              userRepo,
		RefreshTokenRepo:      refreshRepo,
		EmailVerificationRepo: emailRepo,
		SocialRepo:            socialRepo,
		JWT:                   jwt,
		EmailSvc:              emailSvc,
		AuditSvc:              auditSvc,
		PermissionSvc:         permissionSvc,
		SocialCfg:             socialCfg,
	}
}
