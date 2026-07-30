package brokerhandler

import (
	"Broker_backend/services/apigateway/internal/transport/http/dto/brokerdto" //то есть тут я не могу вызвать потому что dto в интернал, а в хендлере не могу вызвать entity потому что entity в интернал? и как правильно сделать?
	"Broker_backend/services/integration/brokerservice/internal/domain/entity"
	brokerv1 "Broker_backend/shared/pkg/grpc/gen/broker/v1"
	"context"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type FixationService interface {
	NewFixation(ctx context.Context, req *brokerdto.FixationRequest, agencyID, userID uuid.UUID) (*entity.FixationCustomer, error)
}

type Handler struct {
	grpc     brokerv1.UnimplementedBrokerServiceServer
	http     http.Server
	fixation FixationService
	logger   *zap.Logger
}

func NewHandler(grpc brokerv1.UnimplementedBrokerServiceServer, http http.Server, fixation FixationService, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{
		grpc:     grpc,
		http:     http,
		fixation: fixation,
		logger:   logger,
	}
}

func (h *Handler) NewFixation(ctx context.Context, req *brokerdto.FixationRequest, agencyID, userID uuid.UUID) (*brokerdto.FixationResponse, error) {
	entityFixation, err := h.fixation.NewFixation(ctx, req, agencyID, userID)
	if err != nil {
		h.logger.Warn("new fixation failed", zap.Error(err))
		return nil, mapHTTPError(err)
	}
	resp := &brokerdto.FixationResponse{
		FixedAt:   entityFixation.FixedAt,
		ExpiresAt: entityFixation.ExpiresAt,
	}

	return resp, nil
}
