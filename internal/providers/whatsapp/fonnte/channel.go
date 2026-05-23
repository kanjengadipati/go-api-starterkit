package fonnte

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"pleco-api/internal/otp"
)

type Channel struct {
	token      string
	baseURL    string
	httpClient *http.Client
}

type sendResponse struct {
	Status  bool   `json:"status"`
	Reason  string `json:"reason"`
	Detail  string `json:"detail"`
	Process string `json:"process"`
}

func New(token, baseURL string, timeout time.Duration) *Channel {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.fonnte.com"
	}
	return &Channel{
		token:      token,
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Channel) SendOTP(ctx context.Context, target string, payload otp.Payload) error {
	if strings.TrimSpace(c.token) == "" {
		return fmt.Errorf("fonnte token is required")
	}

	form := url.Values{}
	form.Set("target", strings.TrimPrefix(target, "+"))
	form.Set("countryCode", "0")
	form.Set("connectOnly", "true")
	form.Set("message", fmt.Sprintf("Your Pleco verification code is:\n\n%s\n\nExpires in %.0f minutes.\n\nDo not share this code.", payload.Code, payload.ExpiresIn.Minutes()))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/send", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if resp.StatusCode >= 400 {
		return fmt.Errorf("fonnte send failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var send sendResponse
	if err := json.Unmarshal(body, &send); err != nil {
		return fmt.Errorf("fonnte send response invalid: %w body=%s", err, string(body))
	}
	if !send.Status {
		reason := firstNonEmpty(send.Reason, send.Detail, "unknown failure")
		return fmt.Errorf("fonnte send failed: %s", reason)
	}

	return nil
}

func (c *Channel) ChannelName() string {
	return "fonnte"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
