// [БОЙЛЕРПЛЕЙТ] Пишется один раз. При новой фиче не трогаешь.
//
// Единственная задача main — запустить и решить, что делать с ошибкой.
// Вся сборка зависимостей живёт в app.Run().

package main

import (
	"fmt"
	"os"

	"Broker_backend/services/brokerservice/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		// Пишем в stderr, а не через zap: логгер мог не создаться.
		fmt.Fprintf(os.Stderr, "brokerservice: %v\n", err)
		os.Exit(1) // ненулевой код — чтобы k8s/docker понял, что упали
	}
}
