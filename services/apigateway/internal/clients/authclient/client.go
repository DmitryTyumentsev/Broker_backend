package authclient

import (
	"Donate_backend/services/authservice/internal/config"
	authv1 "Donate_backend/shared/pkg/grpc/gen/auth/v1"
	"context"
)

type Client struct {
	auth   authv1.AuthServiceClient
	config *config.Config
}

func NewClient(auth authv1.AuthServiceClient, cfg *config.Config) *Client {
	return &Client{auth: auth, config: cfg}
}

func (c *Client) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.TokenPairResponse, error) { //TODO: как чаще на high load проектах в рф делают - хендлер преобразовывает dto в grpcReq или transport в этой функции? наверно логично в хендлере потому что он и нужен для пасринга и преобразований, а транспорт отвечает просто за отправку, так?
	ctx, cancel := context.WithTimeout(ctx, c.config.ContextTimeout)
	defer cancel()

	return c.auth.Register(ctx, req)
}
