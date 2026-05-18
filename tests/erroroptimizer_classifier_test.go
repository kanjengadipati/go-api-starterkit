package tests

import (
	"errors"
	"testing"

	"pleco-api/internal/domain"
	"pleco-api/internal/erroroptimizer"
	"pleco-api/internal/services"
)

func TestDefaultErrorClassifier_Classify(t *testing.T) {
	classifier := &erroroptimizer.DefaultErrorClassifier{}
	endpoint := "/auth/login"

	tests := []struct {
		name     string
		err      error
		wantCode string
	}{
		{
			name:     "Nil Error",
			err:      nil,
			wantCode: "",
		},
		{
			name:     "Weak Password",
			err:      services.ErrWeakPassword,
			wantCode: "AUTH_WEAK_PASSWORD",
		},
		{
			name:     "Invalid Credentials",
			err:      domain.ErrInvalidCredentials,
			wantCode: "AUTH_INVALID_CREDENTIALS",
		},
		{
			name:     "Account Locked",
			err:      domain.ErrAccountLocked,
			wantCode: "AUTH_ACCOUNT_LOCKED",
		},
		{
			name:     "Rate Limit Exceeded",
			err:      errors.New("rate limit exceeded for endpoint"),
			wantCode: "RATE_LIMIT_EXCEEDED",
		},
		{
			name:     "Too Many Requests",
			err:      errors.New("TOO MANY REQUESTS"),
			wantCode: "RATE_LIMIT_EXCEEDED",
		},
		{
			name:     "Email Not Verified",
			err:      domain.ErrEmailNotVerified,
			wantCode: "AUTH_EMAIL_NOT_VERIFIED",
		},
		{
			name:     "Database Error",
			err:      errors.New("pq: database is offline"),
			wantCode: "SERVER_DATABASE_ERROR",
		},
		{
			name:     "Unknown Generic Error",
			err:      errors.New("something very weird happened"),
			wantCode: "SERVER_INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := classifier.Classify(tt.err, endpoint)
			if tt.err == nil {
				if metadata != nil {
					t.Errorf("Classify() expected nil metadata for nil error")
				}
				return
			}
			
			if metadata == nil {
				t.Errorf("Classify() returned nil for error: %v", tt.err)
				return
			}

			if metadata.Code != tt.wantCode {
				t.Errorf("Classify() code = %v, want %v", metadata.Code, tt.wantCode)
			}
		})
	}
}
