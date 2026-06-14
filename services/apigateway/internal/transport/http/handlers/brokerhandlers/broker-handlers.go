package brokerhandlers

import (
	"Broker_backend/services/apigateway/internal/clients/brokerclient"
	"Broker_backend/services/apigateway/internal/transport/http/httperr"
	"fmt"

	validate "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type BrokerHandler struct {
	logger    *zap.Logger
	client    *brokerclient.Client
	validator *validate.Validate
}

func NewBrokerHandler(logger *zap.Logger, client *brokerclient.Client, validator *validate.Validate) *BrokerHandler {
	return &BrokerHandler{
		logger:    logger,
		client:    client,
		validator: validator,
	}
}

func (h *BrokerHandler) FixationCustomer(c *fiber.Ctx) error {
	const op = "handlers.BrokerHandler.FixationCustomer"
	if h == nil {
		return fmt.Errorf("op: %s BrokerHandler is nil", op)
	}

	resp, err := h.client.FixationCustomer(c)
	if err != nil {
		h.logger.Error(err.Error()) //зачем нужен логгер я не понимаю. Это же дублирование ошибки. Куда мы его подключаем, как? ответь подробно максимально - как писать логи, куда их складывать, почему они есть когда есть просто ошибки err и как мне все это потрогать чтоб увидеть реально. Последнее особенно подробно объясни
		return httperr.WriteGRPCError(c, err)
	}

	c.JSON(resp)
}
