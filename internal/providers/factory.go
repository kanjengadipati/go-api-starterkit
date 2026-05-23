package providers

import (
	"strings"
	"time"

	"pleco-api/internal/config"
	"pleco-api/internal/otp"
	emailresend "pleco-api/internal/providers/email/resend"
	emailservice "pleco-api/internal/providers/email/service"
	"pleco-api/internal/providers/noop"
	cloudapi "pleco-api/internal/providers/whatsapp/cloudapi"
	"pleco-api/internal/providers/whatsapp/fonnte"
	"pleco-api/internal/services"
)

func NewOTPChannels(cfg config.AppConfig) map[string]otp.Channel {
	channels := make(map[string]otp.Channel)

	switch strings.ToLower(strings.TrimSpace(cfg.WhatsApp.Provider)) {
	case "fonnte":
		channels["whatsapp"] = fonnte.New(cfg.WhatsApp.FonnteToken, cfg.WhatsApp.FonnteBaseURL, time.Duration(cfg.WhatsApp.TimeoutSeconds)*time.Second)
	case "whatsapp_cloud", "meta", "cloud":
		channels["whatsapp"] = cloudapi.New(cfg.WhatsApp.CloudAccessToken, cfg.WhatsApp.CloudPhoneNumberID, cfg.WhatsApp.CloudAPIBaseURL, cfg.WhatsApp.CloudAPIVersion, time.Duration(cfg.WhatsApp.TimeoutSeconds)*time.Second)
	case "", "disabled":
		channels["whatsapp"] = noop.New("disabled")
	}

	switch strings.ToLower(strings.TrimSpace(cfg.Email.Provider)) {
	case "resend":
		channels["email"] = emailresend.New(cfg.Email.APIKey, cfg.Email.APIBaseURL, cfg.Email.From, cfg.Email.FromName, cfg.Email.ReplyTo, time.Duration(cfg.Email.TimeoutSeconds)*time.Second)
	case "smtp", "sendgrid":
		channels["email"] = emailservice.New(cfg.Email.Provider, services.NewEmailService(cfg.Email))
	case "", "disabled":
		channels["email"] = noop.New("disabled")
	}

	return channels
}
