package service

import (
	"context"

	"pleco-api/internal/otp"
	"pleco-api/internal/services"
)

type Channel struct {
	name  string
	email services.EmailService
}

func New(name string, email services.EmailService) *Channel {
	return &Channel{name: name, email: email}
}

func (c *Channel) SendOTP(_ context.Context, target string, payload otp.Payload) error {
	return c.email.SendOTP(target, payload.Code, payload.ExpiresIn)
}

func (c *Channel) ChannelName() string {
	if c.name == "" {
		return "email"
	}
	return c.name
}
