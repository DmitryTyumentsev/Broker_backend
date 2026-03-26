package validators

import (
	"unicode"
	"unicode/utf8"
)

func IsStrongPassword(password string) bool {

}

func isValidUsername(username string) bool {
	// Для login/публичного имени лучше иметь узкие и понятные правила.
	// Потом будет меньше боли в поиске, индексах, URL и UI.
	if username == "" {
		return false
	}

	runes := utf8.RuneCountInString(username)
	if runes < 3 || runes > 32 {
		return false
	}

	for _, r := range username {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '.':
		default:
			return false
		}
	}

	return true
}

func isStrongEnoughPassword(password string) bool {
	// bcrypt работает максимум с 72 байтами.
	// Лучше отрезать слишком длинные значения сразу на валидации.
	if len(password) == 0 || len(password) > 72 {
		return false
	}

	// По rune, а не по byte — чтобы длина для пользователя
	// считалась более ожидаемо.
	if utf8.RuneCountInString(password) < 8 {
		return false
	}

	var hasLower bool
	var hasUpper bool
	var hasDigit bool

	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	return hasLower && hasUpper && hasDigit
}
