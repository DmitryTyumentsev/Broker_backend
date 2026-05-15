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
} //Почему EncodeToString принимает срез, а не массив(если это так)? почему токен хэшируется через sha256/sha512 с size 32, а пароли через bcrypt с cost 10-12? и к слову cost и size одно и тоже? оба влияют на число итераций? кстати что за итерации это? мы передаем туда каждый раз разные токены, в чем проблема с массивом работать? Что такое итерации в bcrypt? зачем итерации проводит? как это выглядит?  cost нужен для времени, а size для вариативности? тогда самое главное непонятно - почему с паролями bcrypt, а с токенами sha256/sha512? логика в числе символов? типа раз пароли как правило мало весят, вариативность можно подобрать, поэтому замедляем через bcrypt. токены весят стандартно много, поэтому подобрать не смогут, значит делаем 256 вариантов на каждый бит(всего байтов 50). Такая логика или как?

func (s *RefreshTokenService) Hash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
