package validators

import (
	"net/mail"
)

func IsEmail(email string) bool {
	email = normalizeString(email)
	return isValidEmail(email)
}

func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	if addr.Address != email {
		return false
	}
	//parts := strings.Split(email, "@")
	//if len(parts) != 2 {
	//	return false
	//}
	//
	//local := parts[0]
	//domainPart := parts[1]
	//
	//if local == "" || domainPart == "" {
	//	return false
	//}
	//if strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") {
	//	return false
	//}
	//if !strings.Contains(domainPart, ".") {
	//	return false
	//}

	return true
}
