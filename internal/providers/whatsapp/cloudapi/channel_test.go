package cloudapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pleco-api/internal/otp"

	"github.com/stretchr/testify/require"
)

func TestSendOTPSendsCloudAPITextMessage(t *testing.T) {
	var authHeader string
	var path string
	var body map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		path = r.URL.Path
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		_, _ = w.Write([]byte(`{"messages":[{"id":"wamid.test"}]}`))
	}))
	defer server.Close()

	channel := New("token-123", "phone-id-123", server.URL, "v99.0", time.Second)
	err := channel.SendOTP(context.Background(), "+628123456789", otp.Payload{Code: "123456", ExpiresIn: 5 * time.Minute})

	require.NoError(t, err)
	require.Equal(t, "Bearer token-123", authHeader)
	require.Equal(t, "/v99.0/phone-id-123/messages", path)
	require.Equal(t, "whatsapp", body["messaging_product"])
	require.Equal(t, "individual", body["recipient_type"])
	require.Equal(t, "628123456789", body["to"])
	require.Equal(t, "text", body["type"])
	text := body["text"].(map[string]interface{})
	require.Equal(t, false, text["preview_url"])
	require.Contains(t, text["body"], "123456")
}

func TestSendOTPReturnsGraphAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"Invalid OAuth access token.","type":"OAuthException","code":190}}`))
	}))
	defer server.Close()

	channel := New("bad-token", "phone-id-123", server.URL, "v99.0", time.Second)
	err := channel.SendOTP(context.Background(), "+628123456789", otp.Payload{Code: "123456", ExpiresIn: 5 * time.Minute})

	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid OAuth access token")
}

func TestSendOTPRejectsMissingMessageID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"messages":[]}`))
	}))
	defer server.Close()

	channel := New("token-123", "phone-id-123", server.URL, "v99.0", time.Second)
	err := channel.SendOTP(context.Background(), "+628123456789", otp.Payload{Code: "123456", ExpiresIn: 5 * time.Minute})

	require.Error(t, err)
	require.Contains(t, err.Error(), "missing message id")
}
