package passwordhasher

import (
	"crypto"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func Hash(passRaw string) (passHash string, err error) {
	const op = "usecases.Hash()"
	if passHashByte, err := bcrypt.GenerateFromPassword([]byte(passRaw), len([]byte(passRaw))); err != nil {
		return "", fmt.Errorf("op: %s, couldn’t hash, err: %w, passRaw: %s, len: %v", op, err, passRaw, len(passRaw))
	}
	return string(passHashByte), nil
}
