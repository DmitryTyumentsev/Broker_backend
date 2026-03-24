package usecases

import (
	"crypto"
	"fmt"
)

func Hash(passRaw string) (passHash string, err error) {
	const op = "usecase.Hash()"
	if passHash, err = crypto.New(passRaw); err != nil {
		return "", fmt.Errorf("op: %s, couldn’t hash, err: %w, passRaw: %s, len: %v", op, err, passRaw, len(passRaw))
	}
	return passHash, nil
}
