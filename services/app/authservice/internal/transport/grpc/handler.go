package grpc

import (
	usecases2 "Broker_backend/services/app/authservice/internal/usecase"
	"context"

	authv1 "Broker_backend/gen/auth/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	authv1.UnimplementedAuthServiceServer

	service *usecases2.Service
	logger  *zap.Logger
}

func NewHandler(service *usecases2.Service, logger *zap.Logger) *Handler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) Register(
	ctx context.Context,
	req *authv1.RegisterRequest,
) (*authv1.RegisterResponse, error) {
	if h == nil || h.service == nil {
		return nil, status.Error(codes.Unavailable, "auth service is not wired")
	}

	resp, err := h.service.Register(ctx, &usecases2.RegisterRequest{
		Email:       req.Email,
		RawPassword: req.Password,
		LastName:    req.LastName,
		FirstName:   req.FirstName,
		MiddleName:  req.MiddleName,
		DeviceID:    req.DeviceId,
	})
	if err != nil {
		h.logger.Warn("register failed", zap.Error(err))
		return nil, mapError(err)
	}

	return &authv1.RegisterResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresInSec: resp.ExpiresInSec,
	}, nil
}

func (h *Handler) Login(
	ctx context.Context,
	req *authv1.LoginRequest,
) (*authv1.LoginResponse, error) {
	if h == nil || h.service == nil {
		return nil, status.Error(codes.Unavailable, "auth service is not wired")
	}

	resp, err := h.service.Login(ctx, &usecases2.LoginRequest{
		Email:       req.Email,
		RawPassword: req.Password,
		DeviceID:    req.DeviceId,
	})
	if err != nil {
		h.logger.Warn("login failed", zap.Error(err))
		return nil, mapError(err)
	}

	return &authv1.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresInSec: resp.ExpiresInSec,
	}, nil
}

func (h *Handler) Refresh(
	ctx context.Context,
	req *authv1.RefreshRequest,
) (*authv1.RefreshResponse, error) {
	if h == nil || h.service == nil {
		return nil, status.Error(codes.Unavailable, "auth service is not wired")
	}

	resp, err := h.service.Refresh(ctx, &usecases2.RefreshRequest{
		RefreshToken: req.RefreshToken,
		DeviceID:     req.DeviceId,
	})
	if err != nil {
		h.logger.Warn("refresh failed", zap.Error(err))
		return nil, mapError(err)
	}

	return &authv1.RefreshResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresInSec: resp.ExpiresInSec,
	}, nil
}

func (h *Handler) Logout(
	ctx context.Context,
	req *authv1.LogoutRequest,
) (*authv1.LogoutResponse, error) {
	if h == nil || h.service == nil {
		return nil, status.Error(codes.Unavailable, "auth service is not wired")
	}

	// proto LogoutRequest больше не несёт device_id — usecase его всё ещё требует, надо решить, что делать (см. заметку)
	if _, err := h.service.Logout(ctx, &usecases2.LogoutRequest{
		RefreshToken: req.RefreshToken,
	}); err != nil {
		h.logger.Warn("logout failed", zap.Error(err))
		return nil, mapError(err)
	}

	return &authv1.LogoutResponse{}, nil
}
