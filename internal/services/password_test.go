package services

import (
	"errors"
	"testing"
)

func TestValidatePasswordStrengthRequiresDiversity(t *testing.T) {
	weakPasswords := []string{
		"short",
		"lowercase1!",
		"UPPERCASE1!",
		"NoNumber!",
		"NoSymbol1",
		"With Space1!",
	}

	for _, password := range weakPasswords {
		if err := ValidatePasswordStrength(password); !errors.Is(err, ErrWeakPassword) {
			t.Fatalf("expected weak password error for %q, got %v", password, err)
		}
	}
}

func TestValidatePasswordStrengthAcceptsStrongPassword(t *testing.T) {
	if err := ValidatePasswordStrength("Secret123!"); err != nil {
		t.Fatalf("expected strong password to pass, got %v", err)
	}
}
