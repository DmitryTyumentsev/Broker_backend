package auth_grpc

import (
	authv1 "Donate_backend/shared/pkg/grpc/gen/auth/v1"
	"context"
	"time"
)

type Handlers struct {
	auth    authv1.AuthServiceClient
	timeout time.Duration
}

func NewHandler(auth authv1.AuthServiceClient, timeout time.Duration) *Handlers {
	return &Handlers{auth: auth, timeout: timeout}
}

func (h *Handlers) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.TokenPairResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	return h.auth.Register(ctx, req)
}
