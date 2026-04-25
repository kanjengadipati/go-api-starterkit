package token

import "time"

type EmailVerificationToken struct {
	ID        uint
	UserID    uint
	Token     string // SHA-256 hex of the raw token emailed to the user (never store the raw token)
	ExpiresAt time.Time
	CreatedAt time.Time
}
