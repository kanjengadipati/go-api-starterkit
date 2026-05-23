package otp

import (
	"context"
	"time"
)

type Payload struct {
	Code      string
	ExpiresIn time.Duration
}

type Channel interface {
	SendOTP(ctx context.Context, target string, payload Payload) error
	ChannelName() string
}
