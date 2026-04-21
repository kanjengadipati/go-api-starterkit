package config

import (
	"strings"
	"testing"
)

func TestAppConfigValidateAcceptsMinimalValidConfig(t *testing.T) {
	cfg := AppConfig{
		Port:        "8080",
		DatabaseURL: "postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable",
		JWTSecret:   []byte("super-secret-key"),
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
	assertContains(t, message, "JWT_SECRET is required")
	assertContains(t, message, "PORT must be a valid number between 1 and 65535")
}

func TestAppConfigValidateRejectsPartialProviderConfiguration(t *testing.T) {
	cfg := AppConfig{
		Port:        "8080",
		DatabaseURL: "postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable",
		JWTSecret:   []byte("super-secret-key"),
		Email: EmailConfig{
			APIKey: "sg-key",
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
	assertContains(t, message, "SENDGRID_API_KEY and SENDGRID_EMAIL must be set together")
	assertContains(t, message, "FACEBOOK_APP_SECRET is required when FACEBOOK_APP_ID is set")
}

func TestAppConfigValidateRequiresAdminCredentialsWhenSeedingIsEnabled(t *testing.T) {
	cfg := AppConfig{
		Port:          "8080",
		DatabaseURL:   "postgresql://postgres:password@localhost:5432/auth_db?sslmode=disable",
		JWTSecret:     []byte("super-secret-key"),
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

func assertContains(t *testing.T, actual, expected string) {
	t.Helper()
	if !strings.Contains(actual, expected) {
		t.Fatalf("expected %q to contain %q", actual, expected)
	}
}
