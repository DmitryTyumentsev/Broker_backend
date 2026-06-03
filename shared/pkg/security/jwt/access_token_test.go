package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

const testSecret = "test-access-secret-change-me-32-bytes-minimum"

func TestAccessTokenVerifierVerify(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	token := testAccessToken(t, accessTokenPayload{
		Subject:   "user-1",
		DeviceID:  "device-1",
		Role:      "broker_team_member",
		TokenType: AccessTokenKind,
		Issuer:    "authservice",
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		ExpiresAt: now.Add(time.Minute).Unix(),
	})

	verifier, err := NewAccessTokenVerifier(AccessTokenVerifierConfig{
		Secret: testSecret,
		Issuer: "authservice",
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}

	claims, err := verifier.Verify(token)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}

	if claims.UserID != "user-1" {
		t.Fatalf("unexpected user id: %s", claims.UserID)
	}

	if claims.DeviceID != "device-1" {
		t.Fatalf("unexpected device id: %s", claims.DeviceID)
	}

	if claims.Role != "broker_team_member" {
		t.Fatalf("unexpected role: %s", claims.Role)
	}
}

func TestAccessTokenVerifierRejectsExpiredToken(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	token := testAccessToken(t, accessTokenPayload{
		Subject:   "user-1",
		DeviceID:  "device-1",
		Role:      "broker_team_member",
		TokenType: AccessTokenKind,
		Issuer:    "authservice",
		IssuedAt:  now.Add(-time.Hour).Unix(),
		NotBefore: now.Add(-time.Hour).Unix(),
		ExpiresAt: now.Add(-time.Minute).Unix(),
	})

	verifier, err := NewAccessTokenVerifier(AccessTokenVerifierConfig{
		Secret: testSecret,
		Issuer: "authservice",
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}

	if _, err := verifier.Verify(token); err != ErrExpiredToken {
		t.Fatalf("expected expired token error, got %v", err)
	}
}

func TestAccessTokenVerifierRejectsInvalidSignature(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	token := testAccessToken(t, accessTokenPayload{
		Subject:   "user-1",
		DeviceID:  "device-1",
		Role:      "broker_team_member",
		TokenType: AccessTokenKind,
		Issuer:    "authservice",
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		ExpiresAt: now.Add(time.Minute).Unix(),
	}) + "tampered"

	verifier, err := NewAccessTokenVerifier(AccessTokenVerifierConfig{
		Secret: testSecret,
		Issuer: "authservice",
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}

	if _, err := verifier.Verify(token); err != ErrInvalidToken {
		t.Fatalf("expected invalid token error, got %v", err)
	}
}

func testAccessToken(t *testing.T, payload accessTokenPayload) string {
	t.Helper()

	header := accessTokenHeader{
		Alg: SupportedAccessTokenAlg,
		Typ: SupportedAccessTokenType,
	}

	encodedHeader := testEncodeJSON(t, header)
	encodedPayload := testEncodeJSON(t, payload)
	signingInput := encodedHeader + "." + encodedPayload

	mac := hmac.New(sha256.New, []byte(testSecret))
	_, _ = mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + signature
}

func testEncodeJSON(t *testing.T, value any) string {
	t.Helper()

	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal token part: %v", err)
	}

	return base64.RawURLEncoding.EncodeToString(raw)
}
