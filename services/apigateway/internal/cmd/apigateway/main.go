package main

import (
	"Donate_backend/services/apigateway/internal/configs"
	"log"
)

func main() {
	cfg, err := configs.LoadConfig()
	log.Printf("cfg: %v", cfg)
	if err != nil {
		panic(err)
	}

}
