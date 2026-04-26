package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"pleco-api/internal/modules/audit"
	permissionless "pleco-api/internal/modules/social"
	userModule "pleco-api/internal/modules/user"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

type socialProfile struct {
	Email          string
	ProviderUserID string
	Name           string
	Avatar         string
}

type facebookDebugResponse struct {
	Data struct {
		AppID   string `json:"app_id"`
		Type    string `json:"type"`
		IsValid bool   `json:"is_valid"`
		UserID  string `json:"user_id"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type facebookUserResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type appleJWKSet struct {
	Keys []appleJWK `json:"keys"`
}

type appleJWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

func (s *authService) SocialLogin(provider string, token string, deviceID, userAgent, ipAddress string) (*AuthTokens, error) {
	profile, provider, err := s.resolveSocialProfile(provider, token)
	if err != nil {
		return nil, err
	}

	if profile.ProviderUserID == "" {
		return nil, errors.New("invalid provider id")
	}

	if profile.Email == "" {
		return nil, fmt.Errorf("email not available from %s", provider)
	}

	if profile.Name == "" {
		profile.Name = profile.Email
	}

	user, err := s.UserRepo.FindByEmail(profile.Email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		user = &userModule.User{
			Email:      profile.Email,
			Name:       profile.Name,
			Role:       "user",
			IsVerified: true,
		}

		if err := s.UserRepo.Create(user); err != nil {
			return nil, err
		}
	}

	if user.Role == "" {
		user.Role = "user"
		if err := s.UserRepo.Update(user); err != nil {
			return nil, err
		}
	}

	social, err := s.SocialRepo.FindByProvider(provider, profile.ProviderUserID)
	if err != nil {
		return nil, err
	}

	if social != nil {
		if social.UserID != user.ID {
			return nil, errors.New("social account already linked to another user")
		}
	} else {
		newSocial := &permissionless.SocialAccount{
			UserID:         user.ID,
			Provider:       provider,
			ProviderUserID: profile.ProviderUserID,
			AvatarURL:      profile.Avatar,
		}
		if err := s.SocialRepo.Create(newSocial); err != nil {
			return nil, err
		}
	}

	tokens, err := s.issueTokens(user.ID, user.Role, user.AccessTokenVersion, deviceID, userAgent, ipAddress)
	if err != nil {
		return nil, err
	}

	s.AuditSvc.SafeRecord(audit.RecordInput{
		ActorUserID: &user.ID,
		Action:      "social_login",
		Resource:    "auth",
		ResourceID:  &user.ID,
		Status:      "success",
		Description: fmt.Sprintf("%s social login succeeded", provider),
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})

	return tokens, nil
}

func (s *authService) resolveSocialProfile(provider string, token string) (*socialProfile, string, error) {
	normalizedProvider := strings.ToLower(strings.TrimSpace(provider))
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, normalizedProvider, errors.New("social token required")
	}

	switch normalizedProvider {
	case "google":
		profile, err := s.validateGoogleToken(token)
		return profile, normalizedProvider, err
	case "facebook":
		profile, err := s.validateFacebookToken(token)
		return profile, normalizedProvider, err
	case "apple":
		profile, err := s.validateAppleToken(token)
		return profile, normalizedProvider, err
	default:
		return nil, normalizedProvider, errors.New("unsupported provider")
	}
}

func (s *authService) validateGoogleToken(token string) (*socialProfile, error) {
	payload, err := idtoken.Validate(context.Background(), token, s.SocialCfg.GoogleClientID)
	if err != nil {
		return nil, errors.New("invalid google token")
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	avatar, _ := payload.Claims["picture"].(string)
	emailVerified, ok := payload.Claims["email_verified"].(bool)
	if !ok || !emailVerified {
		return nil, errors.New("email not verified by google")
	}

	return &socialProfile{
		Email:          email,
		ProviderUserID: payload.Subject,
		Name:           name,
		Avatar:         avatar,
	}, nil
}

func (s *authService) validateFacebookToken(token string) (*socialProfile, error) {
	if s.SocialCfg.FacebookAppID == "" || s.SocialCfg.FacebookAppSecret == "" {
		return nil, errors.New("facebook social login not configured")
	}

	appToken := s.SocialCfg.FacebookAppID + "|" + s.SocialCfg.FacebookAppSecret
	debugURL := "https://graph.facebook.com/debug_token?" + url.Values{
		"input_token":  {token},
		"access_token": {appToken},
	}.Encode()

	var debugResp facebookDebugResponse
	if err := s.getJSON(debugURL, &debugResp); err != nil {
		return nil, err
	}
	if debugResp.Error != nil {
		return nil, fmt.Errorf("facebook token validation failed: %s", debugResp.Error.Message)
	}
	if !debugResp.Data.IsValid || debugResp.Data.UserID == "" {
		return nil, errors.New("invalid facebook token")
	}
	if debugResp.Data.AppID != s.SocialCfg.FacebookAppID {
		return nil, errors.New("facebook token audience mismatch")
	}
	if debugResp.Data.Type != "" && !strings.EqualFold(debugResp.Data.Type, "USER") {
		return nil, errors.New("unsupported facebook token type")
	}

	userURL := "https://graph.facebook.com/me?" + url.Values{
		"fields":       {"id,name,email,picture.type(large)"},
		"access_token": {token},
	}.Encode()

	var userResp facebookUserResponse
	if err := s.getJSON(userURL, &userResp); err != nil {
		return nil, err
	}
	if userResp.Error != nil {
		return nil, fmt.Errorf("failed to fetch facebook profile: %s", userResp.Error.Message)
	}

	return &socialProfile{
		Email:          userResp.Email,
		ProviderUserID: firstNonEmpty(userResp.ID, debugResp.Data.UserID),
		Name:           userResp.Name,
		Avatar:         userResp.Picture.Data.URL,
	}, nil
}

func (s *authService) validateAppleToken(token string) (*socialProfile, error) {
	if s.SocialCfg.AppleClientID == "" {
		return nil, errors.New("apple social login not configured")
	}

	claims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(incoming *jwt.Token) (interface{}, error) {
		if incoming.Method.Alg() != jwt.SigningMethodES256.Alg() {
			return nil, errors.New("unexpected apple signing method")
		}

		kid, _ := incoming.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("apple token missing kid")
		}

		return s.fetchApplePublicKey(kid)
	})
	if err != nil || !parsedToken.Valid {
		return nil, errors.New("invalid apple token")
	}

	if issuer, _ := claims["iss"].(string); issuer != "https://appleid.apple.com" {
		return nil, errors.New("invalid apple issuer")
	}
	if !claimMatchesAudience(claims["aud"], s.SocialCfg.AppleClientID) {
		return nil, errors.New("apple token audience mismatch")
	}

	email, _ := claims["email"].(string)
	if !claimBool(claims["email_verified"]) {
		return nil, errors.New("email not verified by apple")
	}

	subject, _ := claims["sub"].(string)

	return &socialProfile{
		Email:          email,
		ProviderUserID: subject,
		Name:           email,
	}, nil
}

func (s *authService) fetchApplePublicKey(kid string) (*ecdsa.PublicKey, error) {
	s.appleKeysMutex.RLock()
	cache := s.appleKeysCache
	cacheTime := s.appleKeysCacheTime
	s.appleKeysMutex.RUnlock()

	if cache == nil || time.Since(cacheTime) > 1*time.Hour {
		s.appleKeysMutex.Lock()
		if s.appleKeysCache == nil || time.Since(s.appleKeysCacheTime) > 1*time.Hour {
			var newKeySet appleJWKSet
			if err := s.getJSON("https://appleid.apple.com/auth/keys", &newKeySet); err != nil {
				s.appleKeysMutex.Unlock()
				return nil, err
			}
			s.appleKeysCache = &newKeySet
			s.appleKeysCacheTime = time.Now()
		}
		cache = s.appleKeysCache
		s.appleKeysMutex.Unlock()
	}

	for _, key := range cache.Keys {
		if key.Kid != kid {
			continue
		}
		if key.Kty != "EC" || key.Crv != "P-256" {
			return nil, errors.New("unsupported apple key type")
		}

		xBytes, err := base64.RawURLEncoding.DecodeString(key.X)
		if err != nil {
			return nil, err
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(key.Y)
		if err != nil {
			return nil, err
		}

		return &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}, nil
	}

	return nil, errors.New("matching apple public key not found")
}

func (s *authService) getJSON(rawURL string, target interface{}) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}

	resp, err := s.socialHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("social provider request failed with status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func claimMatchesAudience(raw interface{}, expected string) bool {
	switch value := raw.(type) {
	case string:
		return value == expected
	case []interface{}:
		for _, item := range value {
			if str, ok := item.(string); ok && str == expected {
				return true
			}
		}
	}

	return false
}

func claimBool(raw interface{}) bool {
	switch value := raw.(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(value, "true")
	}

	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}

	return ""
}
