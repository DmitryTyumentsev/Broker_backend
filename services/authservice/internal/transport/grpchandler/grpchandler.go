package grpchandler

import (
	"context"

	"Donate_backend/services/authservice/internal/usecases"
	authv1 "Donate_backend/shared/pkg/grpc/gen/auth/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	authv1.UnimplementedAuthServiceServer

	service *usecases.Service
}

func NewHandler(service *usecases.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.TokenPairResponse, error) {
	if h.service == nil {
		return nil, status.Error(codes.Unavailable, "auth service is not wired yet")
	}

	resp, err := h.service.Register(ctx, &usecases.RegisterRequest{
		Email:       req.Email,
		RawPassword: req.Password,
		LastName:    req.LastName,
		FirstName:   req.FirstName,
		MiddleName:  req.MiddleName,
		DeviceID:    req.DeviceId,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.TokenPairResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresInSec: resp.ExpiresInSec,
	}, nil
}
