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
	bodyDTO, ok := middleware.ValidatedBody[brokerdto.ConnectCustomerRequest](c)
	if ok == false {
		h.logger.Error("middleware.ValidatedBody error: type dto didn't match with c.Locals(validatedBodyKey)")
		return c.JSON(httperr.WriteBadRequest(c, "invalid request"))
	}

	protoDTO := &brokerv1.ConnectCustomerRequest{ //правильно ли вообще передавать managerID, brokerID или можно как-то проще, например просто из принципала вытаскивать? как делают на больших проектах? второй вопрос - как проверять что мне не подставляют чужие данные в запросе? делают ли это в хендлере или миддлварах, есть ли вообще у меня это?
		CustomerID: bodyDTO.CustomerID,
		BrokerID:   bodyDTO.BrokerID,
		ManagerID:  bodyDTO.ManagerID,
	}

	ctx := c.UserContext() //что у нас будет внутри ctx? у нас был *fiber.Ctx в котором конфиги, данные. А в context.Context и то и то уйдет? что в нем будет?
	protoResp, err := h.client.CreateFixationCustomer(ctx, protoDTO)
	if err != nil {
		middleware.AuditLog(
			c,
			h.logger,
			"create fixation customer is failed",
			zap.Error(err), //стоит ли так писать в аудит логах и почему?
			zap.String("customer_id", bodyDTO.CustomerID),
			zap.String("broker_id", bodyDTO.BrokerID),
			zap.String("manager_id", bodyDTO.ManagerID),
		)
		h.logger.Error("client.CreateFixationCustomer error", zap.Error(err)) //какой формат у ошибок в больших проектах в таких ситуациях пишут? что тут писать и зачем если у нас такой подробный access logger? как на больших проектах принято? и второй вопрос - как тут правильнее писать по уровню ошибки - это warning или error? тут же может быть как бизнесово ошибка так и технически. По какому принципу выбираем уровень логирования на ошибку?
		return err
	}
	resp := &brokerdto.ConnectCustomerResponse{ //стоит ли возвращать структуру в таком и подобных кейсах?
		ManagerLastName:   protoResp.ManagerLastName,
		ManagerFirstName:  protoResp.ManagerFirstName,
		ManagerMiddleName: protoResp.ManagerMiddleName,
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}
