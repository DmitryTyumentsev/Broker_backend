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

	"go.uber.org/zap"
)

const (
	supportedAccessTokenAlg  = "HS256"
	supportedAccessTokenType = "JWT"
	accessTokenKind          = "access"

	minAccessTokenSecretLen = 32
)

type AccessTokenIssuer struct {
	secret []byte
	ttl    time.Duration
	issuer string
	logger *zap.Logger
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

type issueAccessTokenInput struct {
	UserID   string
	DeviceID string
	Role     string
	Now      time.Time
}

func NewAccessTokenIssuer(
	cfg *config.Config,
	logger *zap.Logger,
) (*AccessTokenIssuer, error) {
	const op = "jwt.NewAccessTokenIssuer"

	if cfg == nil {
		return nil, fmt.Errorf("%s: config is nil", op)
	}

	if err := validateAccessTokenConfig(cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	secret := strings.TrimSpace(cfg.Business.AccessTokenSecret)

	return &AccessTokenIssuer{
		secret: []byte(secret),
		ttl:    cfg.Business.LifetimeAccessToken,
		issuer: strings.TrimSpace(cfg.Business.AccessTokenIssuer),
		logger: logger,
	}, nil
}

func (i *AccessTokenIssuer) Issue(
	userID string,
	deviceID string,
	role string,
	now time.Time,
) (string, error) {
	const op = "jwt.AccessTokenIssuer.Issue"

	if i == nil {
		return "", fmt.Errorf("%s: issuer is nil", op)
	}

	input := issueAccessTokenInput{
		UserID:   userID,
		DeviceID: deviceID,
		Role:     role,
		Now:      now,
	}

	if err := validateIssueAccessTokenInput(input); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	now = input.Now.UTC()

	header := accessTokenHeader{
		Alg: supportedAccessTokenAlg,
		Typ: supportedAccessTokenType,
	}

	payload := accessTokenPayload{
		Subject:   strings.TrimSpace(input.UserID),
		DeviceID:  strings.TrimSpace(input.DeviceID),
		Role:      strings.TrimSpace(input.Role),
		TokenType: accessTokenKind,
		Issuer:    i.issuer,
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		ExpiresAt: now.Add(i.ttl).Unix(),
	}

	encodedHeader, err := marshalJSONAndEncodeBase64URL(header)
	if err != nil {
		return "", fmt.Errorf("%s: encode header: %w", op, err)
	}

	encodedPayload, err := marshalJSONAndEncodeBase64URL(payload)
	if err != nil {
		return "", fmt.Errorf("%s: encode payload: %w", op, err)
	}

	signingInput := buildSigningInput(encodedHeader, encodedPayload)

	encodedSignature := signHS256AndEncodeBase64URL(signingInput, i.secret)

	token := buildJWT(encodedHeader, encodedPayload, encodedSignature)

	i.logger.Debug(
		"access token issued",
		zap.String("user_id", payload.Subject),
		zap.String("device_id", payload.DeviceID),
		zap.String("role", payload.Role),
		zap.Int64("expires_at", payload.ExpiresAt),
	)

	return token, nil
}

func validateAccessTokenConfig(cfg *config.Config) error {
	alg := strings.TrimSpace(cfg.Business.AccessTokenAlg)
	if alg == "" {
		return errors.New("access token alg is required")
	}

	if alg != supportedAccessTokenAlg {
		return fmt.Errorf("unsupported access token alg: %s", alg)
	}

	typ := strings.TrimSpace(cfg.Business.AccessTokenType)
	if typ == "" {
		return errors.New("access token type is required")
	}

	if typ != supportedAccessTokenType {
		return fmt.Errorf("unsupported access token type: %s", typ)
	}

	secret := strings.TrimSpace(cfg.Business.AccessTokenSecret)
	if secret == "" {
		return errors.New("access token secret is required")
	}

	if len(secret) < minAccessTokenSecretLen {
		return fmt.Errorf("access token secret must be at least %d bytes", minAccessTokenSecretLen)
	}

	if cfg.Business.LifetimeAccessToken <= 0 {
		return errors.New("access token ttl must be positive")
	}

	return nil
}

func validateIssueAccessTokenInput(input issueAccessTokenInput) error {
	if strings.TrimSpace(input.UserID) == "" {
		return errors.New("user id is required")
	}

	if strings.TrimSpace(input.DeviceID) == "" {
		return errors.New("device id is required")
	}

	if strings.TrimSpace(input.Role) == "" {
		return errors.New("role is required")
	}

	if input.Now.IsZero() {
		return errors.New("now must not be zero")
	}

	return nil
}

func marshalJSONAndEncodeBase64URL(v any) (string, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("json marshal: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(jsonBytes), nil
}

func buildSigningInput(encodedHeader string, encodedPayload string) string {
	return encodedHeader + "." + encodedPayload
}

func buildJWT(encodedHeader string, encodedPayload string, encodedSignature string) string {
	return encodedHeader + "." + encodedPayload + "." + encodedSignature
}

func signHS256AndEncodeBase64URL(signingInput string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)

	_, _ = mac.Write([]byte(signingInput))

	signatureBytes := mac.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(signatureBytes)
}

func verifyAccessTokenSignature(token, secret string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}

	signingInput := buildSigningInput(parts[0], parts[1])

	mac := hmac.New(sha256.New, []byte(secret))
	_, err := mac.Write([]byte(signingInput))
	if err != nil {
		return false
	}
	expectedSignature := mac.Sum(nil)

	currentSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}

	return hmac.Equal(currentSignature, expectedSignature)
}
