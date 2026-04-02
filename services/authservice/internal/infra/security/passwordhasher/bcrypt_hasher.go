package passwordhasher

import (
	"Donate_backend/services/authservice/internal/domain"
	"crypto"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher struct {
	//TODO: я верно присоединил интерфейс? моя логика такая:
	// мы завели интерфейс в доменном уровне, который в дальнейшем будет подменять для структуры с тестами например.
	//Для того что метод Hash из passwordHasher пакета работал с этим интерфейсом, достаточно просто создать пустую структуру, которую позже заполним
	//тестовой структурой какой-то
}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{}
}

func (p *PasswordHasher) Hash(passRaw string) (passHash string, err error) {
	const op = "usecases.Hash()"
	if passHashByte, err := bcrypt.GenerateFromPassword([]byte(passRaw), len([]byte(passRaw))); err != nil {
		return "", fmt.Errorf("op: %s, couldn’t hash, err: %w, passRaw: %s, len: %v", op, err, passRaw, len(passRaw))
	}
	return string(passHashByte), nil
}
