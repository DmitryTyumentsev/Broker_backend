package grpc

import (
	fixationv1 "Broker_backend/gen/fixation/v1"
	"context"
	"net/http"

	"go.uber.org/zap"
)

type FixationService interface {
	NewFixation(ctx context.Context, req *fixationv1.NewFixationRequest) (
		*fixationv1.NewFixationResponse, error)
}

type Handler struct {
	grpc     fixationv1.UnimplementedFixationServiceServer
	http     http.Server
	fixation FixationService
	logger   *zap.Logger
}

func NewHandler(grpc fixationv1.UnimplementedFixationServiceServer, fixation FixationService, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{
		grpc:     grpc,
		fixation: fixation,
		logger:   logger,
	}
}

func (h *Handler) NewFixation(ctx context.Context, req *fixationv1.NewFixationRequest) (
	*fixationv1.NewFixationResponse, error) {
	entityFixation, err := h.fixation.NewFixation(ctx, req)
	if err != nil {
		h.logger.Warn("new fixation failed", zap.Error(err))
		return nil, mapGRPCError(err)
	}
	resp := &fixationv1.NewFixationResponse{
		FixedAt:   entityFixation.FixedAt,
		ExpiresAt: entityFixation.ExpiresAt,
	}

	return resp, nil
}
