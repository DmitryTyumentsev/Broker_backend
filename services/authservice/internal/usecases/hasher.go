package usecases

import (
	"crypto"
	"fmt"
) //TODO: где хранить этот файл hasher.go?

func Hash(passRaw string) (passHash string, err error) {
	const op = "usecases.Hash()"
	if passHash, err = crypto.New(passRaw); err != nil {
		return "", fmt.Errorf("op: %s, couldn’t hash, err: %w, passRaw: %s, len: %v", op, err, passRaw, len(passRaw))
	}
	return passHash, nil
}
