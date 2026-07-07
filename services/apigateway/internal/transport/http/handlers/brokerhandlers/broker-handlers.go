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

func (h *BrokerHandler) CreateFixationCustomer(c *fiber.Ctx) error {
	const op = "handlers.brokerhandlers.CreateFixationCustomer" //аудит аксесс логи, трейс еще что-то где пишу? или они вообще каждый вопрос через миддлвары уже фиксируют? и второй вопрос - стоит ли здесь в хендлере использовать op? если да то где?

	bodyDTO, ok := middleware.ValidatedBody[brokerdto.ConnectCustomerRequest](c) //писать if !ok правильно? ведь в ok может записаться true или false и смысл тогда теряется?
	if ok == false {
		h.logger.Error("middleware.ValidatedBody error: type dto didn't match with c.Locals(validatedBodyKey)")
		return c.JSON(httperr.WriteBadRequest(c, "invalid request"))
	}

	protoDTO := &brokerv1.ConnectCustomerRequest{ //правильно ли вообще передавать managerID, brokerID или можно как-то проще, например просто из принципала вытаскивать? как делают на больших проектах? второй вопрос - как проверять что мне не подставляют чужие данные в запросе? делают ли это в хендлере или миддлварах, есть ли вообще у меня это?
		CustomerID: bodyDTO.CustomerID,
		BrokerID:   bodyDTO.BrokerID,
		ManagerID:  bodyDTO.ManagerID,
	}

	ctx := c.UserContext()
	protoResp, err := h.client.CreateFixationCustomer(ctx, protoDTO)
	if err != nil {
		h.logger.Error("client.CreateFixationCustomer error", zap.Error(err)) //какой формат у ошибок в больших проектах в таких ситуациях пишут? стоит ли логировать то что пришло от клиента?
		return err
	}
	resp := &brokerdto.ConnectCustomerResponse{
		ManagerLastName:   protoResp.ManagerLastName,
		ManagerFirstName:  protoResp.ManagerFirstName,
		ManagerMiddleName: protoResp.ManagerMiddleName,
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}
