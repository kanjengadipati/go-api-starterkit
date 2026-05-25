package services

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const PasswordHashCost = 12
const MinimumPasswordLength = 8

var ErrWeakPassword = errors.New("password must be at least 8 characters and include uppercase, lowercase, number, and symbol characters")

func ValidatePasswordStrength(password string) error {
	if utf8.RuneCountInString(password) < MinimumPasswordLength {
		return ErrWeakPassword
	}
	var hasUpper, hasLower, hasNumber, hasSymbol bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char), unicode.IsSymbol(char):
			hasSymbol = true
		}
	}
	if !hasUpper || !hasLower || !hasNumber || !hasSymbol || strings.Contains(password, " ") {
		return ErrWeakPassword
	}
	return nil
}

func HashPassword(password string) (string, error) {
	if err := ValidatePasswordStrength(password); err != nil {
		return "", err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), PasswordHashCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
