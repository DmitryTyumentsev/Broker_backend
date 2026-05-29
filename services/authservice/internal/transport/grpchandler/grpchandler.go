package grpchandler

import (
	"context"

	"Broker_backend/services/authservice/internal/usecases"
	authv1 "Broker_backend/shared/pkg/grpc/gen/auth/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	authv1.UnimplementedAuthServiceServer

	service *usecases.Service
	logger  *zap.Logger
}

func NewHandler(service *usecases.Service, logger *zap.Logger) *Handler {
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
) (*authv1.TokenPairResponse, error) {
	if h == nil || h.service == nil {
		return nil, status.Error(codes.Unavailable, "auth service is not wired")
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
		h.logger.Warn("register failed", zap.Error(err))
		return nil, mapError(err)
	}

	return tokenPairToProto(resp), nil
}

func (h *Handler) Login(
	ctx context.Context,
	req *authv1.LoginRequest,
) (*authv1.TokenPairResponse, error) {
	if h == nil || h.service == nil {
		return nil, status.Error(codes.Unavailable, "auth service is not wired")
	}

	resp, err := h.service.Login(ctx, &usecases.LoginRequest{
		Email:       req.Email,
		RawPassword: req.Password,
		DeviceID:    req.DeviceId,
	})
	if err != nil {
		h.logger.Warn("login failed", zap.Error(err))
		return nil, mapError(err)
	}

	return tokenPairToProto(resp), nil
}

func (h *Handler) Refresh(
	ctx context.Context,
	req *authv1.RefreshRequest,
) (*authv1.TokenPairResponse, error) {
	if h == nil || h.service == nil {
		return nil, status.Error(codes.Unavailable, "auth service is not wired")
	}

	resp, err := h.service.Refresh(ctx, &usecases.RefreshRequest{
		RefreshToken: req.RefreshToken,
		DeviceID:     req.DeviceId,
	})
	if err != nil {
		h.logger.Warn("refresh failed", zap.Error(err))
		return nil, mapError(err)
	}

	return tokenPairToProto(resp), nil
}

func (h *Handler) Logout(
	ctx context.Context,
	req *authv1.RefreshRequest,
) (*authv1.LogoutResponse, error) {
	if h == nil || h.service == nil {
		return nil, status.Error(codes.Unavailable, "auth service is not wired")
	}

	resp, err := h.service.Logout(ctx, &usecases.LogoutRequest{
		RefreshToken: req.RefreshToken,
		DeviceID:     req.DeviceId,
	})
	if err != nil {
		h.logger.Warn("logout failed", zap.Error(err))
		return nil, mapError(err)
	}

	return &authv1.LogoutResponse{
		AllDevice: resp.AllDevice,
		DeviceId:  resp.DeviceID,
	}, nil
}

func tokenPairToProto(resp *usecases.TokenPairResponse) *authv1.TokenPairResponse {
	if resp == nil {
		return &authv1.TokenPairResponse{}
	}

	return &authv1.TokenPairResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresInSec: resp.ExpiresInSec,
	}
}
