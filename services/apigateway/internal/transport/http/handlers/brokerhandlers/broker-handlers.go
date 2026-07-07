package brokerhandlers

import (
	"Broker_backend/services/apigateway/internal/clients/brokerclient"
	"Broker_backend/services/apigateway/internal/transport/http/dto/brokerdto"
	"Broker_backend/services/apigateway/internal/transport/http/httperr"
	"Broker_backend/services/apigateway/internal/transport/http/middleware"
	"errors"
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

func (h *BrokerHandler) Validate() error {
	switch {
	case h == nil:
		return errors.New("broker handler is nil")
	case h.brokerclient == nil:
		return errors.New("broker client is required")
	default:
		return h.brokerclient.Validate()
	}
}

func (h *BrokerHandler) ConnectCustomer(c *fiber.Ctx) error {
	const op = "handlers.BrokerHandler.ConnectCustomer"
	if h == nil {
		return fmt.Errorf("op: %s BrokerHandler is nil", op)
	}

	dto, ok := middleware.ValidatedBody[brokerdto.ConnectCustomerRequest](c) //немного не понимаю зачем мы сначала кладем dto в c.Locals, а затем в хендлере проверяем что в c.Locals dto. Мы же передали дважды(первый раз когда заполняли c.Locals, второй в этой строке), как они могут не совпасть?
	if !ok {
		// и другой вопрос по этой же части - а как избегать дублирования? что мы одно место несколько раз логируем - в апи гейтвей, на границе слоев, в сервисе. Ок ли одновременно логировать и так и так? и главное - зачем вообще есть разделение когда можем логировать одним и раз уж разделили - где что мы смотрим? где отображаются аудит логи, где обычные? где вообще на обычный логгер у нас миддлвар или он не нужен тут?
		h.logger.Error(
			op, //в msg что принято писать - op или суть события(например суть ошибки)?
			zap.String("error", "type dto doesn't match"),
		) //так принято логировать? где trace тогда пишем и остальное? есть ли вообще формат у логеров принятый как писать техническое и как писать аудит лог?
		return httperr.WriteBadRequest(c, "invalid request")
	}

	protoDTO := &brokerv1.ConnectCustomerRequest{
		CustomerID: dto.CustomerID,
		BrokerID:   dto.BrokerID,
		ManagerID:  dto.ManagerID,
	}

	ctx := c.UserContext()
	protoResp, err := h.client.ConnectCustomer(ctx, protoDTO)
	if err != nil {
		h.logger.Error(err.Error()) //зачем нужен логгер я не понимаю. Это же дублирование ошибки. Куда мы его подключаем, как? ответь подробно максимально - как писать логи, куда их складывать, почему они есть когда есть просто ошибки err и как мне все это потрогать чтоб увидеть реально. Последнее особенно подробно объясни
		return httperr.WriteGRPCError(c, err)
	}

	resp := &brokerdto.ConnectCustomerResponse{
		ManagerLastName:   protoResp.ManagerLastName,
		ManagerFirstName:  protoResp.ManagerFirstName,
		ManagerMiddleName: protoResp.ManagerMiddleName,
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}
