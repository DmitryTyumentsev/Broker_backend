package brokerhandlers

import (
	"Broker_backend/services/apigateway/internal/clients/brokerclient"
	"Broker_backend/services/apigateway/internal/transport/http/httperr"
	"Broker_backend/services/apigateway/internal/transport/http/middleware"
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

func (h *BrokerHandler) ConnectCustomer(c *fiber.Ctx) error {
	const op = "handlers.BrokerHandler.ConnectCustomer"
	if h == nil {
		return fmt.Errorf("op: %s BrokerHandler is nil", op)
	}

	dto := c.Locals(middleware.validatedBodyLocalKey)
	ctx := c.UserContext() //я не понимаю где сейчас у меня лежат данные и зачем нужен UserContext. Сейчас картина такая(рассуждаю): мы получаем изначально по сети через файбер байты, файбер кладет их в c(*fiber.Ctx). Далее в миддлварах эти байты (req) мы провалидировали, положили через c.Locals(неясно для чего) в мапу c.Locals и вот нам надо достать данные. Мне надо через c.Locals ключ ввести?
	resp, err := h.client.ConnectCustomer(&ctx, dto)
	if err != nil {
		h.logger.Error(err.Error()) //зачем нужен логгер я не понимаю. Это же дублирование ошибки. Куда мы его подключаем, как? ответь подробно максимально - как писать логи, куда их складывать, почему они есть когда есть просто ошибки err и как мне все это потрогать чтоб увидеть реально. Последнее особенно подробно объясни
		return httperr.WriteGRPCError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}
