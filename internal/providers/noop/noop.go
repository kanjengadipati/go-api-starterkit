package noop

import (
	"context"
	"errors"

	"pleco-api/internal/otp"
)

type Channel struct {
	name string
}

func New(name string) *Channel {
	return &Channel{name: name}
}

func (c *Channel) SendOTP(_ context.Context, _ string, _ otp.Payload) error {
	return errors.New("otp provider is disabled")
}

func (c *Channel) ChannelName() string {
	return c.name
}
