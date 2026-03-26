package validators

import (
	"net/mail"
)

func IsEmail(email string) bool {
	mail.ParseAddress()
}
