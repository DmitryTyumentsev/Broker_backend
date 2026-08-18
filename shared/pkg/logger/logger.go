// Package logger — сборка zap одинаковая для всех сервисов.
//
// Была в трёх копиях, в каждом app.go своя. Копии разъезжаются: настройку
// правят в одном сервисе, а два остальных продолжают шуметь.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New собирает логгер под окружение.
//
// local  — человекочитаемый вывод с цветом
// прочее — JSON, его парсит Loki/ELK
func New(environment string) (*zap.Logger, error) {
	// Стектрейс — только с Error и выше.
	//
	// zap.NewDevelopment() по умолчанию печатает стек начиная с Warn,
	// и это делает лог нечитаемым: «логин не прошёл» и «фиксация уже есть» —
	// штатные исходы бизнес-логики, а не аварии. Двадцать строк стека на
	// каждый неверный пароль прячут те две строки, ради которых лог и читают.
	//
	// Правило: стек нужен там, где никто не предвидел ситуацию.
	// Предвиденный отказ стека не требует.
	stacktrace := zap.AddStacktrace(zapcore.ErrorLevel)

	if environment == "local" {
		return zap.NewDevelopment(stacktrace)
	}

	return zap.NewProduction(stacktrace)
}
