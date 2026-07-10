package validators

import (
	"unicode"
	"unicode/utf8"
)

func IsStrongPassword(password string) bool {
	if utf8.RuneCountInString(password) < 8 {
		return false
	}

	var hasDigit bool
	var hasUpper bool
	var hasLower bool

	for _, r := range password {
		switch {
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		}
	}

	return hasDigit && hasUpper && hasLower
}
