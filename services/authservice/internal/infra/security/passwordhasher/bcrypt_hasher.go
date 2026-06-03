package passwordhasher

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	costBcrypt = 11
)

type PasswordHasher struct {
	cost int
}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		cost: costBcrypt,
	}
}

func NewPasswordHasherWithCost(cost int) *PasswordHasher {
	return &PasswordHasher{
		cost: cost,
	}
}

func (p *PasswordHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	if err != nil {
		return "", fmt.Errorf("generate bcrypt hash: %w", err)
	}

	return string(hash), nil
}

func (p *PasswordHasher) Verify(hashPass string, rawPass string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashPass), []byte(rawPass)) == nil
}
