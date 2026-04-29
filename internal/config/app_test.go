package config

import (
	"strings"
	"testing"
)

func TestAppConfigValidateAcceptsMinimalValidConfig(t *testing.T) {
	cfg := AppConfig{
		Port:        "8080",
		DatabaseURL: "postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable",
		JWTSecret:   []byte("super_secret_key_123_must_be_32_bytes_long_minimum"),
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected config to be valid, got error: %v", err)
	}
}

func TestAppConfigValidateRejectsMissingRequiredValues(t *testing.T) {
	cfg := AppConfig{}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	message := err.Error()
	assertContains(t, message, "DATABASE_URL is required")
	assertContains(t, message, "JWT_SECRET must be at least 32 bytes")
	assertContains(t, message, "PORT must be a valid number between 1 and 65535")
}

func TestAppConfigValidateRejectsPartialProviderConfiguration(t *testing.T) {
	cfg := AppConfig{
		Port:        "8080",
		DatabaseURL: "postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable",
		JWTSecret:   []byte("super_secret_key_123_must_be_32_bytes_long_minimum"),
		Email: EmailConfig{
			Provider: "sendgrid",
			APIKey:   "sg-key",
		},
		Social: SocialConfig{
			FacebookAppID: "fb-app-id",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	message := err.Error()
	assertContains(t, message, "EMAIL_FROM is required when EMAIL_PROVIDER is sendgrid")
	assertContains(t, message, "FACEBOOK_APP_SECRET is required when FACEBOOK_APP_ID is set")
}

func TestAppConfigValidateRejectsUnsupportedEmailProvider(t *testing.T) {
	cfg := AppConfig{
		Port:        "8080",
		DatabaseURL: "postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable",
		JWTSecret:   []byte("super_secret_key_123_must_be_32_bytes_long_minimum"),
		Email: EmailConfig{
			Provider: "mailgun",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	assertContains(t, err.Error(), "EMAIL_PROVIDER must be one of: disabled, sendgrid, resend, smtp")
}

func TestAppConfigValidateRequiresSMTPSettings(t *testing.T) {
	cfg := AppConfig{
		Port:        "8080",
		DatabaseURL: "postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable",
		JWTSecret:   []byte("super_secret_key_123_must_be_32_bytes_long_minimum"),
		Email: EmailConfig{
			Provider: "smtp",
			From:     "noreply@example.com",
			SMTPHost: "",
			SMTPPort: 0,
			SMTPMode: "invalid",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	assertContains(t, err.Error(), "EMAIL_SMTP_HOST is required when EMAIL_PROVIDER is smtp")
	assertContains(t, err.Error(), "EMAIL_SMTP_PORT must be a valid port when EMAIL_PROVIDER is smtp")
	assertContains(t, err.Error(), "EMAIL_SMTP_MODE must be one of: starttls, tls, plain")
}

func TestAppConfigValidateRequiresAdminCredentialsWhenSeedingIsEnabled(t *testing.T) {
	cfg := AppConfig{
		Port:          "8080",
		DatabaseURL:   "postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable",
		JWTSecret:     []byte("super_secret_key_123_must_be_32_bytes_long_minimum"),
		AutoRunSeeds:  true,
		AdminEmail:    "admin@example.com",
		AdminPassword: "",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	assertContains(t, err.Error(), "ADMIN_EMAIL and ADMIN_PASSWORD are required when AUTO_RUN_SEEDS is enabled")
}

func TestAppConfigValidateRejectsUnsupportedAIProvider(t *testing.T) {
	cfg := AppConfig{
		Port:        "8080",
		DatabaseURL: "postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable",
		JWTSecret:   []byte("super_secret_key_123_must_be_32_bytes_long_minimum"),
		AI: AIConfig{
			Enabled:  true,
			Provider: "something-else",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	assertContains(t, err.Error(), "AI_PROVIDER must be one of: mock, ollama, openai, gemini")
}

func TestAppConfigValidateRejectsInvalidAITimeout(t *testing.T) {
	cfg := AppConfig{
		Port:        "8080",
		DatabaseURL: "postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable",
		JWTSecret:   []byte("super_secret_key_123_must_be_32_bytes_long_minimum"),
		AI: AIConfig{
			Enabled:        true,
			Provider:       "mock",
			TimeoutSeconds: 0,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	assertContains(t, err.Error(), "AI_TIMEOUT_SECONDS must be greater than 0 when AI is enabled")
}

func TestAppConfigValidateRequiresAPIKeyForOpenAIAndGemini(t *testing.T) {
	testCases := []struct {
		name     string
		provider string
	}{
		{name: "openai", provider: "openai"},
		{name: "gemini", provider: "gemini"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := AppConfig{
				Port:        "8080",
				DatabaseURL: "postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable",
				JWTSecret:   []byte("super_secret_key_123_must_be_32_bytes_long_minimum"),
				AI: AIConfig{
					Enabled:        true,
					Provider:       tc.provider,
					Model:          "test-model",
					TimeoutSeconds: 30,
				},
			}

			err := cfg.Validate()
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}

			assertContains(t, err.Error(), "AI_API_KEY is required when AI_PROVIDER is "+tc.provider)
		})
	}
}

func assertContains(t *testing.T, actual, expected string) {
	t.Helper()
	if !strings.Contains(actual, expected) {
		t.Fatalf("expected %q to contain %q", actual, expected)
	}
}
