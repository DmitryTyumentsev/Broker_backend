package main

import (
	"Donate_backend/services/authservice/internal/config"
	"Donate_backend/services/authservice/internal/transport/grpchandler"
	"fmt"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalln("config not setup err: %w", err)
	}
	fmt.Printf("config is setup, value: %v", *cfg)

	//logger
	//ctx
	//postgres
	//redis
	//usecases
	//domain
	//transport
	handler := grpchandler.NewHandler(service)
}
