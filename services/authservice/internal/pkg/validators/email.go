package validators

import "net/mail"

func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}

	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Name == "" && addr.Address == email
}
