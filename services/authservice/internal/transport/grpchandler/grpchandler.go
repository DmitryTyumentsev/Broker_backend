package grpchandler

import (
	"Donate_backend/services/authservice/internal/usecases"
	authv1 "Donate_backend/shared/pkg/grpc/gen/auth/v1"
	"context"
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
	reqCMD := &usecases.RegisterRequest{ //TODO: так и принято что cmd структуры лежат в usecase? я вообще не понимаю логику делать отдельно cmd структуры. Тебе что так что при dto надо менять код, не понимаю профит все равно. Пока выглядит как просто лишний кусок кода
		Email:    req.Email,
		Password: req.Password,
		Username: req.Username,
		DeviceID: req.DeviceId,
	}
	res, err := h.service.Register(ctx, reqCMD)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.TokenPairResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiresInSec: res.ExpiresInSec,
	}, nil
}
