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

	"Donate_backend/services/authservice/internal/config"
)

const (
	accessTokenAlg  = "HS256"
	accessTokenType = "JWT"
	accessTokenKind = "access"
)

type AccessTokenIssuer struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

type accessTokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type accessTokenPayload struct {
	UserID    string `json:"sub"`
	DeviceID  string `json:"device_id"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	Issuer    string `json:"iss,omitempty"`
	IssuedAt  int64  `json:"iat"`
	NotBefore int64  `json:"nbf"`
	ExpiresAt int64  `json:"exp"`
}

func NewAccessTokenIssuer(cfg *config.Config) (*AccessTokenIssuer, error) {
	if err := validateAccessTokenIssuerCfg(cfg); err != nil {
		return nil, err
	}

	secret := strings.TrimSpace(cfg.Business.AccessTokenSecret)
	
	if err := validateAccessTokenIssuerSecret(secret); err != nil {
		return nil, err
	}

	return &AccessTokenIssuer{
		secret: []byte(secret),
		ttl:    cfg.Business.LifetimeAccessToken,
		issuer: strings.TrimSpace(cfg.Business.AccessTokenIssuer),
	}, nil
}

func (i *AccessTokenIssuer) Issue(
	userID string,
	deviceID string,
	role string,
	now time.Time,
) (string, error) {
	if i == nil {
		return "", errors.New("access token issuer is nil")
	}

	userID = strings.TrimSpace(userID)
	deviceID = strings.TrimSpace(deviceID)
	role = strings.TrimSpace(role)

	if userID == "" {
		return "", errors.New("user id is required")
	}

	if deviceID == "" {
		return "", errors.New("device id is required")
	}

	if role == "" {
		return "", errors.New("role is required")
	}

	if now.IsZero() {
		now = time.Now().UTC()
	}

	now = now.UTC()

	header := accessTokenHeader{
		Alg: accessTokenAlg,
		Typ: accessTokenType,
	}

	payload := accessTokenPayload{
		UserID:    userID,
		DeviceID:  deviceID,
		Role:      role,
		TokenType: accessTokenKind,
		Issuer:    i.issuer,
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		ExpiresAt: now.Add(i.ttl).Unix(),
	}

	encodedHeader, err := marshalAndEncodeBase64URL(header)
	if err != nil {
		return "", fmt.Errorf("encode access token header: %w", err)
	}

	encodedPayload, err := marshalAndEncodeBase64URL(payload)
	if err != nil {
		return "", fmt.Errorf("encode access token payload: %w", err)
	}

	signingInput := encodedHeader + "." + encodedPayload

	encodedSignature := signHS256(signingInput, i.secret)

	return signingInput + "." + encodedSignature, nil
}

func marshalAndEncodeBase64URL(v any) (string, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("json marshal: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(jsonBytes), nil
}

func signHS256(signingInput string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)

	mac.Write([]byte(signingInput))

	signatureBytes := mac.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(signatureBytes)
}

func validateAccessTokenIssuerCfg(cfg *config.Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}

	if cfg.Business.LifetimeAccessToken <= 0 {
		return errors.New("lifetime access token must be positive")
	}

	if cfg.Business.AccessTokenAlg != "" && cfg.Business.AccessTokenAlg != accessTokenAlg {
		return fmt.Errorf("unsupported access token alg: %s", cfg.Business.AccessTokenAlg)
	}

	if cfg.Business.AccessTokenType != "" && cfg.Business.AccessTokenType != accessTokenType {
		return fmt.Errorf("unsupported access token type: %s", cfg.Business.AccessTokenType)
	}
	return nil
}

func validateAccessTokenIssuerSecret(secret string) error {
	if secret == "" {
		return errors.New("access token secret is required")
	}

	if len(secret) < 32 {
		return errors.New("access token secret must be at least 32 bytes")
	}
	return nil
}
