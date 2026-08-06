package grpc

import (
	fixationv1 "Broker_backend/gen/fixation/v1"
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"Broker_backend/services/integration/fixationservice/internal/usecase"
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FixationService interface {
	NewFixation(ctx context.Context, req *usecase.FixationRequest) (*entity.Fixation, error)
}

type Handler struct {
	grpc     fixationv1.UnimplementedFixationServiceServer
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
	cmdReq, err := convertToUsecaseFixation(req)
	if err != nil {
		return nil, mapGRPCError(err)
	}

	entityFixation, err := h.fixation.NewFixation(ctx, cmdReq)
	if err != nil {
		return nil, mapGRPCError(err)
	}

	resp := &fixationv1.NewFixationResponse{
		FixationId: entityFixation.FixationID.String(),
		FixedAt:    timestamppb.New(entityFixation.FixedAt),
		ExpiresAt:  timestamppb.New(entityFixation.ExpiresAt),
	}

	return resp, nil
}

func convertToUsecaseFixation(req *fixationv1.NewFixationRequest) (*usecase.FixationRequest, error) {
	agencyID, err := uuid.Parse(req.AgencyId)
	if err != nil {
		return nil, err
	}
	fixFor, err := uuid.Parse(req.FixFor)
	if err != nil {
		return nil, err
	}
	fixBy, err := uuid.Parse(req.FixBy)
	if err != nil {
		return nil, err
	}
	projectID, err := uuid.Parse(req.ProjectId)
	if err != nil {
		return nil, err
	}
	return &usecase.FixationRequest{
		AgencyID:  agencyID,
		FixFor:    fixFor,
		FixBy:     fixBy,
		Phone:     req.Phone,
		ProjectID: projectID,
	}, nil
}
