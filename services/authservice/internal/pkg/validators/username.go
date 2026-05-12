package validators

import (
	"unicode"
	"unicode/utf8"
)

func IsValidUsername(username string) bool {
	length := utf8.RuneCountInString(username)
	if length < 3 || length > 32 {
		return false
	}

	for _, r := range username {
		switch {
		case unicode.IsLetter(r):
		case unicode.IsDigit(r):
		case r == '_':
		case r == '-':
		default:
			return false
		}
	}

	return true
}
