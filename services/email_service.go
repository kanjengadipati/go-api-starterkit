package services

import (
	"fmt"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// EmailService defines methods for sending emails (interface)
type EmailService interface {
	SendVerificationEmail(toEmail, token string) error
}

// emailService implements EmailService interface
type emailService struct {
	apiKey string
	from   string
}

var _ EmailService = (*emailService)(nil)

// NewEmailService returns a new emailService as EmailService
func NewEmailService() EmailService {
	return &emailService{
		apiKey: os.Getenv("SENDGRID_API_KEY"),
		from:   os.Getenv("SENDGRID_EMAIL"),
	}
}

func (s *emailService) SendVerificationEmail(toEmail, token string) error {
	link := fmt.Sprintf("http://localhost:8080/verify?token=%s", token)

	from := mail.NewEmail("Go App", s.from)
	subject := "Verify Your Email"
	to := mail.NewEmail("", toEmail)

	plainText := fmt.Sprintf("Click this link to verify: %s", link)
	htmlContent := fmt.Sprintf(`
		<strong>Verify your account</strong><br>
		Click here: <a href="%s">Verify Email</a>
	`, link)

	message := mail.NewSingleEmail(from, subject, to, plainText, htmlContent)

	client := sendgrid.NewSendClient(s.apiKey)
	_, err := client.Send(message)

	return err
}
