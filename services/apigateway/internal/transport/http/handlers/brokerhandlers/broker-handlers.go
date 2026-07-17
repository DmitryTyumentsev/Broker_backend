package brokerhandlers

import (
	"Broker_backend/services/apigateway/internal/clients/brokerclient"
	"Broker_backend/services/apigateway/internal/transport/http/dto/brokerdto"
	"Broker_backend/services/apigateway/internal/transport/http/httperr"
	"Broker_backend/services/apigateway/internal/transport/http/middleware"
	"errors"

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
		return errors.New("brokerservice handler is nil")
	case h.client == nil:
		return errors.New("brokerservice client is required")
	default:
		return h.client.Validate()
	}
}

func (h *BrokerHandler) CreateFixationCustomer(c *fiber.Ctx) error { //Верно понял что согласно моему миддлвару аксесс лог, каждый(вообще каждый) вызов всех методов из цепочки очень подробно записывается и трейсится ещё плюсом?
	bodyDTO, ok := middleware.ValidatedBody[brokerdto.FixationCustomerRequest](c)
	if ok == false {
		h.logger.Error("middleware.ValidatedBody error: type dto didn't match with c.Locals(validatedBodyKey)")
		return c.JSON(httperr.WriteBadRequest(c, "invalid request"))
	}
	principal, ok := middleware.CurrentPrincipal(c) //почитал код, вроде у меня уже есть принципал через middleware.Auth. А как его вытащить, как я сейчас написал? зачем тогда в c клали?
	if !ok {
	}
	fixedBy := principal.UserID
	protoDTO := &brokerv1.FixationCustomerRequest{ //правильно ли вообще передавать managerID, brokerID или можно как-то проще, например просто из принципала вытаскивать? как делают на больших проектах? второй вопрос - как проверять что мне не подставляют чужие данные в запросе? делают ли это в хендлере или миддлварах, есть ли вообще у меня это?
		BrokerID:   bodyDTO.BrokerID,
		CustomerID: bodyDTO.CustomerID,
		FixedBy:    fixedBy,
		FixFor:     bodyDTO.FixFor,
	}

	ctx := c.UserContext() //что у нас будет внутри ctx? у нас был *fiber.Ctx в котором конфиги, данные. А в context.Context и то и то уйдет? что в нем будет?
	protoResp, err := h.client.CreateFixationCustomer(ctx, protoDTO)
	if err != nil {
		middleware.AuditLog(
			c,
			h.logger,
			"create fixation customer is failed",
			zap.Error(err), //стоит ли так писать в аудит логах и почему?
			zap.String("broker_id", bodyDTO.BrokerID),
			zap.String("customer_id", bodyDTO.CustomerID),
			zap.String("fix_for", bodyDTO.FixFor),
		)
		h.logger.Error("client.CreateFixationCustomer error", zap.Error(err)) //какой формат у ошибок в больших проектах в таких ситуациях пишут? что тут писать и зачем если у нас такой подробный access logger? как на больших проектах принято? и второй вопрос - как тут правильнее писать по уровню ошибки - это warning или error? тут же может быть как бизнесово ошибка так и технически. По какому принципу выбираем уровень логирования на ошибку?
		return err
	}
	//почему надо ставить отдельно c.Set("Location", endpoint + ID) ? каждый раз ли это пишут в хендлере отдельно? и по самой логике не очень понял для чего возвращать слово Location и эндпоинт?
	resp := &brokerdto.FixationCustomerResponse{ //мы возвращаем отдельно dto вместо напрямую protoResp потому что в dto есть json теги, а в protoResp нет? а если добавить?
		FixationID: protoResp.FixationID,
		Status:     protoResp.Status,
		FixedAt:    protoResp.FixedAt,
		ExpiresAt:  protoResp.ExpiresAt,
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}
