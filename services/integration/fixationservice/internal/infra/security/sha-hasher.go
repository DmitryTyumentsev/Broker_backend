package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

func SignHS256AndEncodeBase64URL(hashSecret string, v string) (string, error) {
	secretBytes := []byte(hashSecret)

	mac := hmac.New(sha256.New, secretBytes)

	_, _ = mac.Write([]byte(v))

	signatureBytes := mac.Sum(nil)

	return base64.StdEncoding.EncodeToString(signatureBytes), nil
}
