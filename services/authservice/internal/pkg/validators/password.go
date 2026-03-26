package validators

import (
	"unicode"
	"unicode/utf8"
)

func IsStrongPassword(password string) bool {
	if len(password) > 72 {
		return false
	}

	if utf8.RuneCountInString(password) < 8 {
		return false
	}

	var hasLower bool
	var hasUpper bool
	var hasDigit bool

	for _, r := range password { //TODO: почему password перевелся в руны? как работает uft8.RuneCountInString? расскажи структурно
		// какие есть кодировки самые основные, чем отличаются, в какой из кодировок лежат руны, чем отличаются руны от байтов,
		//почему руны это int32, сколько рун в английском, русском языках, символы типа _-=+.?/"';:,[]{}&%$#@!><   .
		//Второй вопрос -как работает IsDigit? это и есть проверка на английкий язык?
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
