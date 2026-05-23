package cloudapi

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
	accessToken   string
	phoneNumberID string
	baseURL       string
	apiVersion    string
	httpClient    *http.Client
}

type textMessageRequest struct {
	MessagingProduct string      `json:"messaging_product"`
	RecipientType    string      `json:"recipient_type"`
	To               string      `json:"to"`
	Type             string      `json:"type"`
	Text             textPayload `json:"text"`
}

type textPayload struct {
	PreviewURL bool   `json:"preview_url"`
	Body       string `json:"body"`
}

type sendResponse struct {
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
	Error *graphError `json:"error"`
}

type graphError struct {
	Message      string `json:"message"`
	Type         string `json:"type"`
	Code         int    `json:"code"`
	ErrorSubcode int    `json:"error_subcode"`
	FBTraceID    string `json:"fbtrace_id"`
}

func New(accessToken, phoneNumberID, baseURL, apiVersion string, timeout time.Duration) *Channel {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://graph.facebook.com"
	}
	if strings.TrimSpace(apiVersion) == "" {
		apiVersion = "v20.0"
	}
	return &Channel{
		accessToken:   accessToken,
		phoneNumberID: phoneNumberID,
		baseURL:       strings.TrimRight(baseURL, "/"),
		apiVersion:    strings.Trim(strings.TrimSpace(apiVersion), "/"),
		httpClient:    &http.Client{Timeout: timeout},
	}
}

func (c *Channel) SendOTP(ctx context.Context, target string, payload otp.Payload) error {
	if strings.TrimSpace(c.accessToken) == "" {
		return fmt.Errorf("whatsapp cloud access token is required")
	}
	if strings.TrimSpace(c.phoneNumberID) == "" {
		return fmt.Errorf("whatsapp cloud phone number id is required")
	}

	reqBody := textMessageRequest{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               strings.TrimPrefix(target, "+"),
		Type:             "text",
		Text: textPayload{
			PreviewURL: false,
			Body:       fmt.Sprintf("Your Pleco verification code is:\n\n%s\n\nExpires in %.0f minutes.\n\nDo not share this code.", payload.Code, payload.ExpiresIn.Minutes()),
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/%s/%s/messages", c.baseURL, c.apiVersion, c.phoneNumberID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	var parsed sendResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return fmt.Errorf("whatsapp cloud response invalid: %w body=%s", err, string(responseBody))
	}
	if parsed.Error != nil {
		return fmt.Errorf("whatsapp cloud send failed: %s", parsed.Error.Message)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("whatsapp cloud send failed: status=%d body=%s", resp.StatusCode, string(responseBody))
	}
	if len(parsed.Messages) == 0 || strings.TrimSpace(parsed.Messages[0].ID) == "" {
		return fmt.Errorf("whatsapp cloud send failed: missing message id")
	}

	return nil
}

func (c *Channel) ChannelName() string {
	return "whatsapp_cloud"
}
