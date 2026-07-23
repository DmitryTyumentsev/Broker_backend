package brokerhandler

import (
	"Broker_backend/services/brokerservice/internal/domain/entity"
	"Broker_backend/services/brokerservice/internal/usecases/cmd"
	brokerv1 "Broker_backend/shared/pkg/grpc/gen/broker/v1"
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service interface {
	NewFixationCustomer(ctx context.Context, dtoFixationCustomer *cmd.FixationCustomerRequest) (*entity.FixationCustomer, error)
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
func (h *Handler) NewFixationCustomer(ctx context.Context, req *brokerv1.NewFixationCustomerRequest,
) (*brokerv1.NewFixationCustomerResponse, error) {
	if h == nil || h.service == nil { //пишут ли проверки ресивера на nil в самом методе как здесь я написал? разве на уровне этого слоя не вешают валидаторы тоже?
		return nil, status.Error(codes.Unavailable, "broker service is not wired")
	}

	cmdFixationCustomer := &cmd.FixationCustomerRequest{ //как принято называть их - дто или cmd или command struct? обрати внимание что это сам сервис а не апи гейтвей
		BrokerID:   req.BrokerId,
		CustomerID: req.CustomerId,
		FixFor:     req.FixFor,
		FixedBy:    req.FixedBy,
	}
	entityFixationCustomer, err := h.service.NewFixationCustomer(ctx, cmdFixationCustomer)
	if err != nil {
		h.logger.Warn("create customer fixation failed", zap.Error(err)) //почему логируем тут? где принято и как логировать правильно сервисные и инфра ошибки?
		return nil, mapError(err)
	}
	resp := &brokerv1.NewFixationCustomerResponse{
		FixationId: entityFixationCustomer.FixationID.String(),
		Status:     brokerv1.FixationStatus.Enum(),
		FixedAt:    timestamppb.New(entityFixationCustomer.FixedAt),
		ExpiresAt:  timestamppb.New(entityFixationCustomer.ExpiresAt),
	}

	return resp, nil
}
