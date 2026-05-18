package config

import (
	"fmt"
	"strconv"
	"strings"
)

type EmailConfig struct {
	Provider       string
	APIKey         string
	APIBaseURL     string
	From           string
	FromName       string
	ReplyTo        string
	TimeoutSeconds int
	SMTPHost       string
	SMTPPort       int
	SMTPUsername   string
	SMTPPassword   string
	SMTPMode       string
	AppBaseURL     string
	FrontendURL    string
}

type SocialProviderConfig struct {
	ClientID     string
	ClientSecret string
}

type SocialConfig struct {
	ActiveProviders []string
	Providers       map[string]SocialProviderConfig
}

type AIConfig struct {
	Enabled        bool
	Provider       string
	Model          string
	BaseURL        string
	APIKey         string
	TimeoutSeconds int
}

type ErrorOptimizationConfig struct {
	Enabled bool
}

type AppConfig struct {
	Port                     string
	DatabaseURL              string
	RedisURL                 string
	TrustedProxies           []string
	CORSAllowedOrigins       []string
	JWTSecret                []byte
	AccessTokenExpiryMinutes int
	AdminEmail               string
	AdminPassword            string
	AutoRunMigrations        bool
	AutoRunSeeds             bool
	Email                    EmailConfig
	Social                   SocialConfig
	AI                       AIConfig
	ErrorOptimization        ErrorOptimizationConfig
}

func LoadAppConfig() AppConfig {
	return AppConfig{
		Port:                     GetEnv("PORT", "8080"),
		DatabaseURL:              GetEnv("DATABASE_URL", ""),
		RedisURL:                 GetEnv("REDIS_URL", ""),
		TrustedProxies:           envList("TRUSTED_PROXIES", []string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}),
		CORSAllowedOrigins:       corsAllowedOrigins(),
		JWTSecret:                []byte(GetEnv("JWT_SECRET", "")),
		AccessTokenExpiryMinutes: envInt("ACCESS_TOKEN_EXPIRY_MINUTES", 15),
		AdminEmail:               GetEnv("ADMIN_EMAIL", ""),
		AdminPassword:            GetEnv("ADMIN_PASSWORD", ""),
		AutoRunMigrations:        envBool("AUTO_RUN_MIGRATIONS"),
		AutoRunSeeds:             envBool("AUTO_RUN_SEEDS"),
		Email: EmailConfig{
			Provider:       strings.ToLower(GetEnv("EMAIL_PROVIDER", "disabled")),
			APIKey:         GetEnv("EMAIL_API_KEY", ""),
			APIBaseURL:     GetEnv("EMAIL_API_BASE_URL", ""),
			From:           GetEnv("EMAIL_FROM", ""),
			FromName:       GetEnv("EMAIL_FROM_NAME", "Go App"),
			ReplyTo:        GetEnv("EMAIL_REPLY_TO", ""),
			TimeoutSeconds: envInt("EMAIL_TIMEOUT_SECONDS", 15),
			SMTPHost:       GetEnv("EMAIL_SMTP_HOST", ""),
			SMTPPort:       envInt("EMAIL_SMTP_PORT", 587),
			SMTPUsername:   GetEnv("EMAIL_SMTP_USERNAME", ""),
			SMTPPassword:   GetEnv("EMAIL_SMTP_PASSWORD", ""),
			SMTPMode:       strings.ToLower(GetEnv("EMAIL_SMTP_MODE", "starttls")),
			AppBaseURL:     firstNonEmptyEnv("APP_BASE_URL", "RENDER_EXTERNAL_URL", "http://localhost:8080"),
			FrontendURL:    GetEnv("FRONTEND_URL", ""),
		},
		Social: loadSocialConfig(),
		AI: AIConfig{
			Enabled:        envBool("AI_ENABLED"),
			Provider:       strings.ToLower(GetEnv("AI_PROVIDER", "mock")),
			Model:          GetEnv("AI_MODEL", "qwen2.5:3b"),
			BaseURL:        GetEnv("AI_BASE_URL", ""),
			APIKey:         GetEnv("AI_API_KEY", ""),
			TimeoutSeconds: envInt("AI_TIMEOUT_SECONDS", 30),
		},
		ErrorOptimization: ErrorOptimizationConfig{
			Enabled: envBool("ENABLE_ERROR_OPTIMIZATION"),
		},
	}
}

func (c AppConfig) Validate() error {
	var problems []string

	if c.DatabaseURL == "" {
		problems = append(problems, "DATABASE_URL is required")
	}

	if len(c.JWTSecret) < 32 {
		problems = append(problems, "JWT_SECRET must be at least 32 bytes for cryptographic strength")
	}

	port, err := strconv.Atoi(strings.TrimSpace(c.Port))
	if err != nil || port < 1 || port > 65535 {
		problems = append(problems, "PORT must be a valid number between 1 and 65535")
	}

	switch c.Email.Provider {
	case "", "disabled":
	case "sendgrid", "resend":
		if c.Email.APIKey == "" {
			problems = append(problems, "EMAIL_API_KEY is required when EMAIL_PROVIDER is "+c.Email.Provider)
		}
		if c.Email.From == "" {
			problems = append(problems, "EMAIL_FROM is required when EMAIL_PROVIDER is "+c.Email.Provider)
		}
	case "smtp":
		if c.Email.From == "" {
			problems = append(problems, "EMAIL_FROM is required when EMAIL_PROVIDER is smtp")
		}
		if c.Email.SMTPHost == "" {
			problems = append(problems, "EMAIL_SMTP_HOST is required when EMAIL_PROVIDER is smtp")
		}
		if c.Email.SMTPPort < 1 || c.Email.SMTPPort > 65535 {
			problems = append(problems, "EMAIL_SMTP_PORT must be a valid port when EMAIL_PROVIDER is smtp")
		}
		switch c.Email.SMTPMode {
		case "starttls", "tls", "plain":
		default:
			problems = append(problems, "EMAIL_SMTP_MODE must be one of: starttls, tls, plain")
		}
	default:
		problems = append(problems, "EMAIL_PROVIDER must be one of: disabled, sendgrid, resend, smtp")
	}

	if c.Email.Provider != "" && c.Email.Provider != "disabled" && c.Email.TimeoutSeconds < 1 {
		problems = append(problems, "EMAIL_TIMEOUT_SECONDS must be greater than 0 when email is enabled")
	}

	for _, p := range c.Social.ActiveProviders {
		providerCfg := c.Social.Providers[p]
		if providerCfg.ClientID == "" {
			problems = append(problems, fmt.Sprintf("SOCIAL_%s_CLIENT_ID is required because %s is an active social provider", strings.ToUpper(p), p))
		}
		if p == "facebook" && providerCfg.ClientSecret == "" {
			problems = append(problems, "SOCIAL_FACEBOOK_CLIENT_SECRET is required because facebook is an active social provider")
		}
	}

	if c.AutoRunSeeds && (c.AdminEmail == "" || c.AdminPassword == "") {
		problems = append(problems, "ADMIN_EMAIL and ADMIN_PASSWORD are required when AUTO_RUN_SEEDS is enabled")
	}

	if c.AI.Enabled {
		if c.AI.TimeoutSeconds < 1 {
			problems = append(problems, "AI_TIMEOUT_SECONDS must be greater than 0 when AI is enabled")
		}

		switch c.AI.Provider {
		case "mock":
		case "ollama":
			if c.AI.BaseURL == "" {
				problems = append(problems, "AI_BASE_URL is required when AI_PROVIDER is ollama")
			}
			if c.AI.Model == "" {
				problems = append(problems, "AI_MODEL is required when AI_PROVIDER is ollama")
			}
		case "openai":
			if c.AI.APIKey == "" {
				problems = append(problems, "AI_API_KEY is required when AI_PROVIDER is openai")
			}
			if c.AI.Model == "" {
				problems = append(problems, "AI_MODEL is required when AI_PROVIDER is openai")
			}
		case "gemini":
			if c.AI.APIKey == "" {
				problems = append(problems, "AI_API_KEY is required when AI_PROVIDER is gemini")
			}
			if c.AI.Model == "" {
				problems = append(problems, "AI_MODEL is required when AI_PROVIDER is gemini")
			}
		case "anthropic":
			if c.AI.APIKey == "" {
				problems = append(problems, "AI_API_KEY is required when AI_PROVIDER is anthropic")
			}
			if c.AI.Model == "" {
				problems = append(problems, "AI_MODEL is required when AI_PROVIDER is anthropic")
			}
		default:
			problems = append(problems, "AI_PROVIDER must be one of: mock, ollama, openai, gemini, anthropic")
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("invalid app config:\n- %s", strings.Join(problems, "\n- "))
	}

	return nil
}

func envBool(key string) bool {
	value := strings.TrimSpace(strings.ToLower(GetEnv(key, "")))
	return value == "1" || value == "true" || value == "yes"
}

func envList(key string, fallback []string) []string {
	value := strings.TrimSpace(GetEnv(key, ""))
	if value == "" {
		return append([]string(nil), fallback...)
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	if len(items) == 0 {
		return append([]string(nil), fallback...)
	}
	return items
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(GetEnv(key, ""))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func firstNonEmptyEnv(keys ...string) string {
	last := ""
	for _, key := range keys {
		if strings.Contains(key, "://") {
			last = key
			continue
		}
		if value := GetEnv(key, ""); value != "" {
			return value
		}
	}
	return last
}

func corsAllowedOrigins() []string {
	defaults := []string{
		"http://localhost:3000",
		"http://127.0.0.1:3000",
	}

	if frontendURL := strings.TrimSpace(GetEnv("FRONTEND_URL", "")); frontendURL != "" {
		defaults = append([]string{frontendURL}, defaults...)
	}

	return envList("CORS_ALLOWED_ORIGINS", defaults)
}

func loadSocialConfig() SocialConfig {
	activeProviders := envList("SOCIAL_ACTIVE_PROVIDERS", nil)
	providers := make(map[string]SocialProviderConfig)

	for _, p := range activeProviders {
		p = strings.ToLower(p)
		prefix := "SOCIAL_" + strings.ToUpper(p) + "_"
		providers[p] = SocialProviderConfig{
			ClientID:     GetEnv(prefix+"CLIENT_ID", ""),
			ClientSecret: GetEnv(prefix+"CLIENT_SECRET", ""),
		}
	}

	return SocialConfig{
		ActiveProviders: activeProviders,
		Providers:       providers,
	}
}
