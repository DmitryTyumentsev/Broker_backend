package brokerhandlers

import (
	"Broker_backend/services/apigateway/internal/clients/brokerclient"
	"Broker_backend/services/apigateway/internal/transport/http/dto/brokerdto"
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

	dto, ok := middleware.ValidatedBody[brokerdto.ConnectCustomerRequest](c) //немного не понимаю зачем мы сначала кладем dto в c.Locals, а затем в хендлере проверяем что в c.Locals dto. Мы же передали дважды(первый раз когда заполняли c.Locals, второй в этой строке), как они могут не совпасть?
	if !ok {
		//тут думаю нужно залогировать. audit или обычный логгер тут использовать? а на сервисном уровне кого из них использовать? аудит логер бизнесовый/событийный, а обычный логер ставится на границах слоев(из того что помню). Тут аудит правильнее? и другой вопрос по этой же части - а как избегать дублирования? что мы одно место несколько раз логируем - в апи гейтвей, на границе слоев, в сервисе. Ок ли одновременно логировать и так и так? и главное - зачем вообще есть разделение когда можем логировать одним и раз уж разделили - где что мы смотрим? где отображаются аудит логи, где обычные? где вообще на обычный логгер у нас миддлвар или он не нужен тут?
		return httperr.WriteBadRequest(c, "invalid request")
	}

	ctx := c.UserContext() //не понимаю зачем через c.UserContext создаем context.Context, а не через context.Background() ? это из-за принципала?
	resp, err := h.client.ConnectCustomer(ctx, &dto)
	if err != nil {
		h.logger.Error(err.Error()) //зачем нужен логгер я не понимаю. Это же дублирование ошибки. Куда мы его подключаем, как? ответь подробно максимально - как писать логи, куда их складывать, почему они есть когда есть просто ошибки err и как мне все это потрогать чтоб увидеть реально. Последнее особенно подробно объясни
		return httperr.WriteGRPCError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}
