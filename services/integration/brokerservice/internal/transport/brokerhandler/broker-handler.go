package brokerhandler

import (
	"Broker_backend/services/integration/brokerservice/internal/domain/entity"
	"Broker_backend/services/integration/brokerservice/internal/usecases"
	brokerv1 "Broker_backend/shared/pkg/grpc/gen/broker/v1"
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service interface {
	NewFixationCustomer(ctx context.Context, dtoFixationCustomer *usecases.FixationCustomerRequest) (*entity.FixationCustomer, error)
	UpdateFixation(ctx context.Context, req *brokerdto.FixationRequest, brokerID, userID uuid.UUID) (*entity.FixationCustomer, error)
}
type Handler struct {
	grpc    brokerv1.UnimplementedBrokerServiceServer
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

	cmdFixationCustomer := &usecases.FixationCustomerRequest{ //как принято называть их - дто или cmd или command struct? обрати внимание что это сам сервис а не апи гейтвей
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
		FixationId: entityFixationCustomer.FixationIDNew.String(),
		Status:     fixationStatusToProto(entityFixationCustomer.StatusActive),
		FixedAt:    timestamppb.New(entityFixationCustomer.FixedAt),
		ExpiresAt:  timestamppb.New(entityFixationCustomer.ExpiresAt),
	}

	return resp, nil
}

func fixationStatusToProto(status entity.Status) brokerv1.FixationStatus {
	switch status {
	case entity.StatusActive:
		return brokerv1.FixationStatus_FIXATION_STATUS_ACTIVE
	case entity.StatusConverted:
		return brokerv1.FixationStatus_FIXATION_STATUS_CONVERTED
	case entity.StatusExpired:
		return brokerv1.FixationStatus_FIXATION_STATUS_EXPIRED
	case entity.StatusRemoved:
		return brokerv1.FixationStatus_FIXATION_STATUS_REMOVED
	default:
		return brokerv1.FixationStatus_FIXATION_STATUS_UNSPECIFIED
	}
}

func (h *Handler) UpdateFixation(ctx context.Context, req *brokerdto.FixationRequest, brokerID, userID uuid.UUID,
) (*brokerdto.FixationResponse, error) {

	entityFixation, err := h.service.UpdateFixation(ctx, req, brokerID, userID)
	if err != nil {
		return nil, mapError(err)
	}
	resp := &brokerdto.FixationResponse{
		FixationID: entityFixation.FixationIDNew,
		Status:     entityFixation.StatusActive,
		FixedAt:    entityFixation.FixedAt,
		ExpiresAt:  entityFixation.ExpiresAt,
	}

	return resp, nil
}
