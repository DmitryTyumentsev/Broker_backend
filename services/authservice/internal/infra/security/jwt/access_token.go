package jwt

import (
	"Donate_backend/services/authservice/internal/config"
	"Donate_backend/services/authservice/internal/domain/entity"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"time"
)

type AccessTokenIssuer struct {
	cfg *config.Config
}

func NewAccessTokenIssuer() *AccessTokenIssuer {
	return &AccessTokenIssuer{}
}

func (i *AccessTokenIssuer) Issue(userID, deviceID, role, tokenType string, now time.Time) (string, error) {
	if i == nil {
		return "", errors.New("AccessTokenIssuer is nil")
	}
	header := &entity.Header{
		AccessTokenAlg:  i.cfg.Business.AccessTokenAlg, //добавляется ли между + и строками пробел в такой записи? userID + deviceID + role + tokenType
		AccessTokenType: i.cfg.Business.AccessTokenType,
	}
	payload := &entity.Payload{
		UserID:    userID,
		DeviceID:  deviceID,
		Role:      role,
		TokenType: tokenType,
		CreatedAt: now,
		ExpiresAt: now.Add(i.cfg.Business.LifetimeAccessToken),
	}
	accessTokenBytes := json.NewDecoder(header).Decode(sha256.New())

	accessToken := sha256.Sum256([]byte(header))
	return "", errors.New("access token issuer is not implemented yet")
}
