package services

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/quotedprintable"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"pleco-api/internal/config"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// EmailService defines methods for sending emails (interface)
type EmailService interface {
	SendVerificationEmail(toEmail, token string) error
	SendPasswordReset(toEmail, token string) error
}

type emailService struct {
	provider    string
	apiKey      string
	apiBaseURL  string
	from        string
	fromName    string
	replyTo     string
	appBaseURL  string
	frontendURL string
	httpClient  *http.Client
	smtpHost    string
	smtpPort    int
	smtpUsername string
	smtpPassword string
	smtpMode    string
}

var _ EmailService = (*emailService)(nil)

func NewEmailService(cfg config.EmailConfig) EmailService {
	return &emailService{
		provider:    strings.ToLower(strings.TrimSpace(cfg.Provider)),
		apiKey:      cfg.APIKey,
		apiBaseURL:  cfg.APIBaseURL,
		from:        cfg.From,
		fromName:    firstNonEmpty(cfg.FromName, "Go App"),
		replyTo:     cfg.ReplyTo,
		appBaseURL:  firstNonEmpty(cfg.AppBaseURL, "http://localhost:8080"),
		frontendURL: cfg.FrontendURL,
		httpClient:  &http.Client{Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second},
		smtpHost:    cfg.SMTPHost,
		smtpPort:    cfg.SMTPPort,
		smtpUsername: cfg.SMTPUsername,
		smtpPassword: cfg.SMTPPassword,
		smtpMode:    strings.ToLower(strings.TrimSpace(cfg.SMTPMode)),
	}
}

func (s *emailService) SendVerificationEmail(toEmail, token string) error {
	verifyBaseURL := firstNonEmpty(s.frontendURL, s.appBaseURL)
	link := fmt.Sprintf("%s/verify?token=%s", trimTrailingSlash(verifyBaseURL), token)

	plainText := fmt.Sprintf("Click this link to verify: %s", link)
	htmlContent := fmt.Sprintf(`
		<strong>Verify your account</strong><br>
		Click here: <a href="%s">Verify Email</a>
	`, link)

	return s.sendEmail(toEmail, "Verify Your Email", plainText, htmlContent)
}

func (s *emailService) SendPasswordReset(toEmail string, token string) error {
	resetBaseURL := firstNonEmpty(s.frontendURL, s.appBaseURL)
	link := fmt.Sprintf("%s/reset-password?token=%s", trimTrailingSlash(resetBaseURL), token)

	plainText := fmt.Sprintf(
		"We received a password reset request for your account. If you did not make this request, you can ignore this email.\n\nTo reset your password, please visit the following link:\n%s\n\nThis link will expire in 15 minutes.",
		link,
	)
	htmlContent := fmt.Sprintf(`
		<strong>Password reset request</strong><br>
		We received a password reset request for your account.<br>
		If you did not make this request, you can ignore this email.<br><br>
		To reset your password, please click the following link:<br>
		<a href="%s">Reset Password</a><br><br>
		This link will expire in 15 minutes.
	`, link)

	return s.sendEmail(toEmail, "Password Reset Request", plainText, htmlContent)
}

func (s *emailService) sendEmail(toEmail, subject, plainText, htmlContent string) error {
	switch s.provider {
	case "", "disabled":
		return nil
	case "sendgrid":
		return s.sendWithSendGrid(toEmail, subject, plainText, htmlContent)
	case "resend":
		return s.sendWithResend(toEmail, subject, plainText, htmlContent)
	case "smtp":
		return s.sendWithSMTP(toEmail, subject, plainText, htmlContent)
	default:
		return fmt.Errorf("unsupported email provider: %s", s.provider)
	}
}

func (s *emailService) sendWithSendGrid(toEmail, subject, plainText, htmlContent string) error {
	from := mail.NewEmail(s.fromName, s.from)
	to := mail.NewEmail("", toEmail)
	message := mail.NewSingleEmail(from, subject, to, plainText, htmlContent)
	client := sendgrid.NewSendClient(s.apiKey)
	response, err := client.Send(message)
	if err != nil {
		return err
	}

	if response != nil && response.StatusCode >= 400 {
		return fmt.Errorf("sendgrid email failed: status=%d, body=%s", response.StatusCode, response.Body)
	}
	return nil
}

func (s *emailService) sendWithResend(toEmail, subject, plainText, htmlContent string) error {
	payload := map[string]interface{}{
		"from":    formatFromHeader(s.fromName, s.from),
		"to":      []string{toEmail},
		"subject": subject,
		"text":    plainText,
		"html":    htmlContent,
	}
	if s.replyTo != "" {
		payload["reply_to"] = s.replyTo
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	endpoint := firstNonEmpty(s.apiBaseURL, "https://api.resend.com") + "/emails"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend email failed: status=%d, body=%s", resp.StatusCode, string(responseBody))
	}

	return nil
}

func (s *emailService) sendWithSMTP(toEmail, subject, plainText, htmlContent string) error {
	addr := net.JoinHostPort(s.smtpHost, fmt.Sprintf("%d", s.smtpPort))

	var (
		client *smtp.Client
		conn   net.Conn
		err    error
	)

	switch s.smtpMode {
	case "tls":
		conn, err = tls.Dial("tcp", addr, &tls.Config{
			ServerName: s.smtpHost,
			MinVersion: tls.VersionTLS12,
		})
		if err != nil {
			return err
		}
		client, err = smtp.NewClient(conn, s.smtpHost)
	default:
		client, err = smtp.Dial(addr)
		if err == nil && s.smtpMode == "starttls" {
			if ok, _ := client.Extension("STARTTLS"); ok {
				if err = client.StartTLS(&tls.Config{
					ServerName: s.smtpHost,
					MinVersion: tls.VersionTLS12,
				}); err != nil {
					_ = client.Close()
					return err
				}
			}
		}
	}
	if err != nil {
		return err
	}
	defer client.Close()

	if s.smtpUsername != "" {
		auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(s.from); err != nil {
		return err
	}
	if err := client.Rcpt(toEmail); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}

	message, err := buildSMTPMessage(formatFromHeader(s.fromName, s.from), toEmail, s.replyTo, subject, plainText, htmlContent)
	if err != nil {
		_ = writer.Close()
		return err
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return client.Quit()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func trimTrailingSlash(value string) string {
	return strings.TrimRight(value, "/")
}

func formatFromHeader(name, email string) string {
	if strings.TrimSpace(name) == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}

func buildSMTPMessage(from, to, replyTo, subject, plainText, htmlContent string) ([]byte, error) {
	boundary := fmt.Sprintf("pleco-boundary-%d", time.Now().UnixNano())

	var body strings.Builder
	body.WriteString(fmt.Sprintf("From: %s\r\n", from))
	body.WriteString(fmt.Sprintf("To: %s\r\n", to))
	if strings.TrimSpace(replyTo) != "" {
		body.WriteString(fmt.Sprintf("Reply-To: %s\r\n", replyTo))
	}
	body.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%q\r\n", boundary))
	body.WriteString("\r\n")
	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	body.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	body.WriteString(encodeQuotedPrintable(plainText))
	body.WriteString("\r\n")
	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	body.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	body.WriteString(encodeQuotedPrintable(htmlContent))
	body.WriteString("\r\n")
	body.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return []byte(body.String()), nil
}

func encodeQuotedPrintable(value string) string {
	var buf bytes.Buffer
	writer := quotedprintable.NewWriter(&buf)
	_, _ = writer.Write([]byte(value))
	_ = writer.Close()
	return buf.String()
}
