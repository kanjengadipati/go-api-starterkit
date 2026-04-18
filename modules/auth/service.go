package auth

import (
	"go-api-starterkit/config"
	"go-api-starterkit/modules/audit"
	permissionless "go-api-starterkit/modules/social"
	tokenModule "go-api-starterkit/modules/token"
	userModule "go-api-starterkit/modules/user"
	"go-api-starterkit/services"

	"gorm.io/gorm"
)

type AuthService interface {
	Register(user *userModule.User, password string) error
	Login(email, password, deviceID, userAgent, ipAddress string) (*AuthTokens, error)
	Logout(userID uint, deviceID string) error
	RefreshToken(oldRefreshToken string) (*AuthTokens, error)
	GetProfile(userID uint) (*userModule.User, error)
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
}

var _ AuthService = (*authService)(nil)

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
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
	}
}
