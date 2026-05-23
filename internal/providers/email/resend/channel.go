package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"pleco-api/internal/otp"
)

type Channel struct {
	apiKey     string
	baseURL    string
	from       string
	fromName   string
	replyTo    string
	httpClient *http.Client
}

func New(apiKey, baseURL, from, fromName, replyTo string, timeout time.Duration) *Channel {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.resend.com"
	}
	return &Channel{
		apiKey:     apiKey,
		baseURL:    strings.TrimRight(baseURL, "/"),
		from:       from,
		fromName:   fromName,
		replyTo:    replyTo,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Channel) SendOTP(ctx context.Context, target string, payload otp.Payload) error {
	if strings.TrimSpace(c.apiKey) == "" {
		return fmt.Errorf("resend api key is required")
	}
	if strings.TrimSpace(c.from) == "" {
		return fmt.Errorf("email from is required")
	}

	plainText := fmt.Sprintf("Your verification code is:\n\n%s\n\nExpires in %.0f minutes.\n\nDo not share this code.", payload.Code, payload.ExpiresIn.Minutes())
	htmlContent := fmt.Sprintf("<p>Your verification code is:</p><p><strong style=\"font-size:24px;letter-spacing:6px\">%s</strong></p><p>Expires in %.0f minutes.</p><p>Do not share this code.</p>", payload.Code, payload.ExpiresIn.Minutes())

	body, err := json.Marshal(map[string]interface{}{
		"from":     formatFrom(c.fromName, c.from),
		"to":       []string{target},
		"subject":  "Your Pleco Verification Code",
		"text":     plainText,
		"html":     htmlContent,
		"reply_to": optionalReplyTo(c.replyTo),
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("resend otp send failed: status=%d body=%s", resp.StatusCode, string(responseBody))
	}

	return nil
}

func (c *Channel) ChannelName() string {
	return "resend"
}

func formatFrom(name, email string) string {
	if strings.TrimSpace(name) == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", name, email)
}

func optionalReplyTo(replyTo string) interface{} {
	if strings.TrimSpace(replyTo) == "" {
		return nil
	}
	return replyTo
}
