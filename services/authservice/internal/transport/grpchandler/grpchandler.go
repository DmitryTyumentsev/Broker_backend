package grpchandler

import (
	authv1 "Donate_backend/shared/pkg/grpc/gen/auth/v1"
	"context"
)

type Handler struct {
	authv1.UnimplementedAuthServiceServer
	app *usecase.Service
}

func NewHandler(app *usecase.Service) *Handler {
	return &Handler{
		app: app}
}

func (h *Handler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.TokenPairResponse, error) {
	reqCMD := usecase.RegisterRequest{ //TODO: есть метод у которого есть ресивер. Допустим как в этой строчке я хочу его вызвать в каком-то
		// другом методе. Вызывать нужно по пакету или по ресиверу? функции понятно что по пакетам а методы у которых есть ресивер?
		//всегда по ресиверу или есть исключения когда надо по пакету?
		Email:    req.Email,
		Pass:     req.Pass,
		Username: req.Username,
		DeviceID: req.DeviceID,
	}
	res, err := h.app.Register(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.TokenPairResponse{
		Access:       res.Access,
		Refresh:      res.Refresh,
		ExpiresInSec: res.ExpiresInSec,
	}, nil
}
