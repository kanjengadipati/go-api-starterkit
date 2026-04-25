package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	Secret []byte
}

func NewJWTService(secret []byte) *JWTService {
	return &JWTService{Secret: secret}
}

func (j *JWTService) GenerateToken(userID uint, role string, duration time.Duration, tokenType string, accessTokenVersion uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"type":    tokenType,
		"tv":      accessTokenVersion,
		"exp":     time.Now().Add(duration).Unix(),
	})

	return token.SignedString(j.Secret)
}

func (j *JWTService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate that the signing method is HMAC (HS256)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.Secret, nil
	}, jwt.WithLeeway(10*time.Second)) // allow small clock drift

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}

func (j *JWTService) GenerateCustomClaimsToken(claims map[string]interface{}, duration time.Duration) (string, error) {
	// Copy claims so we don't mutate input map
	cpy := jwt.MapClaims{}
	for k, v := range claims {
		cpy[k] = v
	}
	cpy["exp"] = time.Now().Add(duration).Unix()
	cpy["iat"] = time.Now().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, cpy)
	return token.SignedString(j.Secret)
}
