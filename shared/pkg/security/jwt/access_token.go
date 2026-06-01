package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	SupportedAccessTokenAlg  = "HS256"
	SupportedAccessTokenType = "JWT"
	AccessTokenKind          = "access"
)

type AccessTokenClaims struct {
	UserID    string
	DeviceID  string
	Role      string
	TokenType string
	Issuer    string
	IssuedAt  time.Time
	NotBefore time.Time
	ExpiresAt time.Time
}

type AccessTokenVerifier struct {
	secret []byte
	issuer string
	now    func() time.Time
}

type AccessTokenVerifierConfig struct {
	Secret string
	Issuer string
	Now    func() time.Time
}

type accessTokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type accessTokenPayload struct {
	Subject   string `json:"sub"`
	DeviceID  string `json:"device_id"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	Issuer    string `json:"iss,omitempty"`
	IssuedAt  int64  `json:"iat"`
	NotBefore int64  `json:"nbf"`
	ExpiresAt int64  `json:"exp"`
}

var (
	ErrMalformedToken = errors.New("malformed token")
	ErrInvalidToken   = errors.New("invalid token")
	ErrExpiredToken   = errors.New("expired token")
	ErrTokenNotActive = errors.New("token is not active")
)

func NewAccessTokenVerifier(cfg AccessTokenVerifierConfig) (*AccessTokenVerifier, error) {
	secret := strings.TrimSpace(cfg.Secret)
	if len(secret) < 32 {
		return nil, fmt.Errorf("access token secret must be at least 32 bytes")
	}

	now := cfg.Now
	if now == nil {
		now = time.Now
	}

	return &AccessTokenVerifier{
		secret: []byte(secret),
		issuer: strings.TrimSpace(cfg.Issuer),
		now:    now,
	}, nil
}

func (v *AccessTokenVerifier) Verify(rawToken string) (AccessTokenClaims, error) {
	if v == nil {
		return AccessTokenClaims{}, fmt.Errorf("access token verifier is nil")
	}

	parts := strings.Split(strings.TrimSpace(rawToken), ".")
	if len(parts) != 3 {
		return AccessTokenClaims{}, ErrMalformedToken
	}

	header, err := decodeSegment[accessTokenHeader](parts[0])
	if err != nil {
		return AccessTokenClaims{}, fmt.Errorf("%w: decode header", ErrMalformedToken)
	}

	if header.Alg != SupportedAccessTokenAlg || header.Typ != SupportedAccessTokenType {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	if !v.validSignature(parts) {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	payload, err := decodeSegment[accessTokenPayload](parts[1])
	if err != nil {
		return AccessTokenClaims{}, fmt.Errorf("%w: decode payload", ErrMalformedToken)
	}

	claims := payloadToClaims(payload)
	if err := v.validateClaims(claims); err != nil {
		return AccessTokenClaims{}, err
	}

	return claims, nil
}

func (v *AccessTokenVerifier) validSignature(parts []string) bool {
	signingInput := parts[0] + "." + parts[1]

	mac := hmac.New(sha256.New, v.secret)
	_, _ = mac.Write([]byte(signingInput))
	expected := mac.Sum(nil)

	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}

	return hmac.Equal(actual, expected)
}

func (v *AccessTokenVerifier) validateClaims(claims AccessTokenClaims) error {
	now := v.now().UTC()

	if strings.TrimSpace(claims.UserID) == "" ||
		strings.TrimSpace(claims.DeviceID) == "" ||
		strings.TrimSpace(claims.Role) == "" ||
		claims.TokenType != AccessTokenKind {
		return ErrInvalidToken
	}

	if v.issuer != "" && claims.Issuer != v.issuer {
		return ErrInvalidToken
	}

	if !claims.NotBefore.IsZero() && now.Before(claims.NotBefore) {
		return ErrTokenNotActive
	}

	if claims.ExpiresAt.IsZero() || !now.Before(claims.ExpiresAt) {
		return ErrExpiredToken
	}

	return nil
}

func payloadToClaims(payload accessTokenPayload) AccessTokenClaims {
	return AccessTokenClaims{
		UserID:    strings.TrimSpace(payload.Subject),
		DeviceID:  strings.TrimSpace(payload.DeviceID),
		Role:      strings.TrimSpace(payload.Role),
		TokenType: strings.TrimSpace(payload.TokenType),
		Issuer:    strings.TrimSpace(payload.Issuer),
		IssuedAt:  unixToTime(payload.IssuedAt),
		NotBefore: unixToTime(payload.NotBefore),
		ExpiresAt: unixToTime(payload.ExpiresAt),
	}
}

func unixToTime(ts int64) time.Time {
	if ts <= 0 {
		return time.Time{}
	}

	return time.Unix(ts, 0).UTC()
}

func decodeSegment[T any](segment string) (T, error) {
	var out T

	raw, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return out, err
	}

	if err := json.Unmarshal(raw, &out); err != nil {
		return out, err
	}

	return out, nil
}
