package main

import (
	"Donate_backend/services/auth-service/internal/config"
	"fmt"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalln("config not setup err: %w", err)
	}
	fmt.Printf("config is setup, value: %v", cfg)

	//ctx := context.Background()
}
