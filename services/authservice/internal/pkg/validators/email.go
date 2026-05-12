package validators

import "net/mail"

func IsValidEmail(email string) bool {
	if email == "" {
		return false
	}

	_, err := mail.ParseAddress(email)
	return err == nil
}
