package main

import (
	"Donate_backend/services/apigateway/internal/config"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	log.Printf("cfg: %v", cfg)
	if err != nil {
		panic(err)
	}

}
