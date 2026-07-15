package brokerhandler

import (
	"Broker_backend/services/brokerservice/internal/domain/entity"
	"Broker_backend/services/brokerservice/internal/usecases/cmd"
	brokerv1 "Broker_backend/shared/pkg/grpc/gen/broker/v1"
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface { //зачем нужен интерфейс? вопрос про именно этот кейс - зачем тут добавили интерфейс? что это дает? почему просто не указать service структурой в domain/usecases(не знаю где правильнее)?
	CreateFixationCustomer(ctx context.Context, brokerID entity.BrokerID, managerID entity.ManagerID, customerID entity.CustomerID) error //допустим метод успешно отработал, я поменял менеджера. как принято на проектах - достаточно просто error в apigateway возвращать? не нужно же managerID возвращать или фио менеджера например?
}
type Handler struct {
	brokerv1.UnimplementedBrokerServiceServer
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{
		service: service,
		logger:  logger,
	}
}
func (h *Handler) CreateFixationCustomer(ctx context.Context, req *brokerv1.ConnectCustomerRequest,
) error {
	if h == nil || h.service == nil { //пишут ли проверки ресивера на nil в самом методе как здесь я написал? разве на уровне этого слоя не вешают валидаторы тоже?
		return status.Error(codes.Unavailable, "broker service is not wired")
	}

	cmdReq := cmd.ConnectCustomerRequest{
		CustomerID: req.CustomerId,
		BrokerID:   req.BrokerId,
		ManagerID:  req.ManagerId,
	}
	err := h.service.CreateFixationCustomer(ctx, cmdReq.BrokerID, cmdReq.ManagerID, cmdReq.CustomerID)
	if err != nil {
		h.logger.Warn("create customer fixation failed", zap.Error(err)) //почему логируем тут? где принято и как логировать правильно сервисные и инфра ошибки?
		return mapError(err)
	}

	return nil
}
