package amocrm

import (
	"Broker_backend/services/integration/fixationservice/internal/config"
	"net/http"
)

type Sender struct {
	Config *config.Config
	HTTP   *http.Client
}

func NewSender(config *config.Config) *Sender {
	return &Sender{
		Config: config,
		HTTP:   &http.Client{},
	}
}
