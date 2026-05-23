package fonnte

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pleco-api/internal/otp"

	"github.com/stretchr/testify/require"
)

func TestSendOTPSendsFonnteForm(t *testing.T) {
	var gotAuth string
	var gotTarget string
	var gotCountryCode string
	var gotConnectOnly string
	var gotMessage string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/send", r.URL.Path)
		require.Equal(t, "token-123", r.Header.Get("Authorization"))
		require.NoError(t, r.ParseForm())
		gotAuth = r.Header.Get("Authorization")
		gotTarget = r.Form.Get("target")
		gotCountryCode = r.Form.Get("countryCode")
		gotConnectOnly = r.Form.Get("connectOnly")
		gotMessage = r.Form.Get("message")
		_, _ = w.Write([]byte(`{"status":true,"detail":"success! message in queue","process":"pending"}`))
	}))
	defer server.Close()

	channel := New("token-123", server.URL, time.Second)
	err := channel.SendOTP(context.Background(), "+628123456789", otp.Payload{Code: "123456", ExpiresIn: 5 * time.Minute})

	require.NoError(t, err)
	require.Equal(t, "token-123", gotAuth)
	require.Equal(t, "628123456789", gotTarget)
	require.Equal(t, "0", gotCountryCode)
	require.Equal(t, "true", gotConnectOnly)
	require.Contains(t, gotMessage, "123456")
}

func TestSendOTPReturnsFonnteReasonOnStatusFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":false,"reason":"target invalid"}`))
	}))
	defer server.Close()

	channel := New("token-123", server.URL, time.Second)
	err := channel.SendOTP(context.Background(), "+628123456789", otp.Payload{Code: "123456", ExpiresIn: 5 * time.Minute})

	require.Error(t, err)
	require.Contains(t, err.Error(), "target invalid")
}

func TestSendOTPReturnsHTTPErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad token", http.StatusUnauthorized)
	}))
	defer server.Close()

	channel := New("token-123", server.URL, time.Second)
	err := channel.SendOTP(context.Background(), "+628123456789", otp.Payload{Code: "123456", ExpiresIn: 5 * time.Minute})

	require.Error(t, err)
	require.Contains(t, err.Error(), "status=401")
	require.Contains(t, err.Error(), "bad token")
}

func TestSendOTPRejectsInvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer server.Close()

	channel := New("token-123", server.URL, time.Second)
	err := channel.SendOTP(context.Background(), "+628123456789", otp.Payload{Code: "123456", ExpiresIn: 5 * time.Minute})

	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "response invalid"))
}
