package validators

import "unicode/utf8"

func IsUsername(username string) bool {
	username = normalizeString(username)
	return isValidUsername(username)
}

func isValidUsername(username string) bool {
	runes := utf8.RuneCountInString(username)
	if runes < 3 || runes > 32 {
		return false
	}

	for _, r := range username {
		switch {
		case r >= 'a' && r <= 'z': //TODO: то есть можно не писать после условия в case ничего кроме : ?
		case r >= '0' && r <= '9':
		case r == '_':
		default:
			return false
		}
	}

	return true
}
