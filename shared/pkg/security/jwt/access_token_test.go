package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

const testSecret = "test-access-secret-change-me-32-bytes-minimum"

var (
	testAgencyID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testUserID   = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

func TestAccessTokenVerifierVerify(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	token := testAccessToken(t, accessTokenPayload{
		AgencyID:  testAgencyID.String(),
		Subject:   testUserID.String(),
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

	if claims.AgencyID != testAgencyID {
		t.Fatalf("unexpected agency id: %s", claims.AgencyID)
	}

	if claims.UserID != testUserID {
		t.Fatalf("unexpected user id: %s", claims.UserID)
	}

	if claims.DeviceID != "device-1" {
		t.Fatalf("unexpected device id: %s", claims.DeviceID)
	}

	if claims.Role != "broker_team_member" {
		t.Fatalf("unexpected role: %s", claims.Role)
	}
}

// Регрессия: sub, который не парсится в uuid, обязан обнулить именно UserID,
// а не соседнее поле. Иначе токен с мусорным sub проходит проверку
// и в принципала уезжает нулевой agency_id при живом user_id.
func TestAccessTokenVerifierRejectsMalformedSubject(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	token := testAccessToken(t, accessTokenPayload{
		AgencyID:  testAgencyID.String(),
		Subject:   "not-a-uuid",
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

	if _, err := verifier.Verify(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected invalid token error, got %v", err)
	}
}

func TestAccessTokenVerifierRejectsExpiredToken(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	token := testAccessToken(t, accessTokenPayload{
		AgencyID:  testAgencyID.String(),
		Subject:   testUserID.String(),
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

	if _, err := verifier.Verify(token); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("expected expired token error, got %v", err)
	}
}

func TestAccessTokenVerifierRejectsInvalidSignature(t *testing.T) {
	now := time.Unix(1_800_000_000, 0).UTC()
	token := testAccessToken(t, accessTokenPayload{
		AgencyID:  testAgencyID.String(),
		Subject:   testUserID.String(),
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

	if _, err := verifier.Verify(token); !errors.Is(err, ErrInvalidToken) {
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
