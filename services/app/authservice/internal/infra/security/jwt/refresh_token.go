package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const defaultRefreshTokenSize = 32

type RefreshTokenService struct {
	size int
}

func NewRefreshTokenService() *RefreshTokenService {
	return &RefreshTokenService{
		size: defaultRefreshTokenSize,
	}
}

func (s *RefreshTokenService) New() (string, error) {
	size := s.size
	if size <= 0 {
		size = defaultRefreshTokenSize
	}

	raw := make([]byte, size)

	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func (s *RefreshTokenService) Hash(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
