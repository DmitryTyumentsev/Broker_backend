package security

import (
	"Broker_backend/services/integration/fixationservice/internal/config"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
)

func SignHS256AndEncodeBase64URL(cfg *config.Config, v string) (string, error) {
	secret, err := json.Marshal(cfg.Business.AccessTokenSecret)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, secret)

	_, _ = mac.Write([]byte(v))

	signatureBytes := mac.Sum(nil)

	return base64.StdEncoding.EncodeToString(signatureBytes), nil
}
