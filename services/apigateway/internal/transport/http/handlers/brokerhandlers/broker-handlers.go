package brokerhandlers

import (
	"Broker_backend/services/apigateway/internal/clients/brokerclient"
	"Broker_backend/services/apigateway/internal/config"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type BrokerHandler struct {
	logger *zap.Logger
	cfg    *config.Config
	client *brokerclient.Client
}

func NewBrokerHandler(logger *zap.Logger, cfg *config.Config, client *brokerclient.Client) *BrokerHandler {
	return &BrokerHandler{
		logger: logger,
		cfg:    cfg,
		client: client,
	}
}

func (h *BrokerHandler) CreateCustomer(ctx *fiber.Ctx) {}
