package config

import (
	"fmt"
	"strconv"
	"strings"
)

type EmailConfig struct {
	APIKey      string
	From        string
	AppBaseURL  string
	FrontendURL string
}

type SocialConfig struct {
	GoogleClientID    string
	FacebookAppID     string
	FacebookAppSecret string
	AppleClientID     string
}

type AIConfig struct {
	Enabled        bool
	Provider       string
	Model          string
	BaseURL        string
	APIKey         string
	TimeoutSeconds int
}

type AppConfig struct {
	Port              string
	DatabaseURL       string
	TrustedProxies    []string
	JWTSecret         []byte
	AdminEmail        string
	AdminPassword     string
	AutoRunMigrations bool
	AutoRunSeeds      bool
	Email             EmailConfig
	Social            SocialConfig
	AI                AIConfig
}

func LoadAppConfig() AppConfig {
	return AppConfig{
		Port:              GetEnv("PORT", "8080"),
		DatabaseURL:       GetEnv("DATABASE_URL", ""),
		TrustedProxies:    envList("TRUSTED_PROXIES", []string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}),
		JWTSecret:         []byte(GetEnv("JWT_SECRET", "")),
		AdminEmail:        GetEnv("ADMIN_EMAIL", ""),
		AdminPassword:     GetEnv("ADMIN_PASSWORD", ""),
		AutoRunMigrations: envBool("AUTO_RUN_MIGRATIONS"),
		AutoRunSeeds:      envBool("AUTO_RUN_SEEDS"),
		Email: EmailConfig{
			APIKey:      GetEnv("SENDGRID_API_KEY", ""),
			From:        GetEnv("SENDGRID_EMAIL", ""),
			AppBaseURL:  firstNonEmptyEnv("APP_BASE_URL", "RENDER_EXTERNAL_URL", "http://localhost:8080"),
			FrontendURL: GetEnv("FRONTEND_URL", ""),
		},
		Social: SocialConfig{
			GoogleClientID:    GetEnv("GOOGLE_CLIENT_ID", ""),
			FacebookAppID:     GetEnv("FACEBOOK_APP_ID", ""),
			FacebookAppSecret: GetEnv("FACEBOOK_APP_SECRET", ""),
			AppleClientID:     GetEnv("APPLE_CLIENT_ID", ""),
		},
		AI: AIConfig{
			Enabled:        envBool("AI_ENABLED"),
			Provider:       strings.ToLower(GetEnv("AI_PROVIDER", "mock")),
			Model:          GetEnv("AI_MODEL", "qwen2.5:3b"),
			BaseURL:        GetEnv("AI_BASE_URL", "http://localhost:11434"),
			APIKey:         GetEnv("AI_API_KEY", ""),
			TimeoutSeconds: envInt("AI_TIMEOUT_SECONDS", 30),
		},
	}
}

func (c AppConfig) Validate() error {
	var problems []string

	if c.DatabaseURL == "" {
		problems = append(problems, "DATABASE_URL is required")
	}

	if len(c.JWTSecret) == 0 {
		problems = append(problems, "JWT_SECRET is required")
	}

	port, err := strconv.Atoi(strings.TrimSpace(c.Port))
	if err != nil || port < 1 || port > 65535 {
		problems = append(problems, "PORT must be a valid number between 1 and 65535")
	}

	if (c.Email.APIKey == "") != (c.Email.From == "") {
		problems = append(problems, "SENDGRID_API_KEY and SENDGRID_EMAIL must be set together")
	}

	if c.Social.FacebookAppID != "" && c.Social.FacebookAppSecret == "" {
		problems = append(problems, "FACEBOOK_APP_SECRET is required when FACEBOOK_APP_ID is set")
	}

	if c.Social.FacebookAppSecret != "" && c.Social.FacebookAppID == "" {
		problems = append(problems, "FACEBOOK_APP_ID is required when FACEBOOK_APP_SECRET is set")
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
		default:
			problems = append(problems, "AI_PROVIDER must be one of: mock, ollama")
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
