package authclient

import (
	"context"
	"fmt"

	"Broker_backend/services/apigateway/internal/config"
	authv1 "Broker_backend/shared/pkg/grpc/gen/auth/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	auth   authv1.AuthServiceClient
	config *config.Config
}

func NewAuthServiceClient(
	ctx context.Context,
	cfg *config.Config,
) (*grpc.ClientConn, authv1.AuthServiceClient, error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("config is nil")
	}

	_ = ctx

	conn, err := grpc.NewClient(
		"passthrough:///"+cfg.AuthGRPC.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create auth service client: %w", err)
	}

	return conn, authv1.NewAuthServiceClient(conn), nil
}

func NewClient(auth authv1.AuthServiceClient, cfg *config.Config) *Client {
	return &Client{
		auth:   auth,
		config: cfg,
	}
}

func (c *Client) Register(
	ctx context.Context,
	req *authv1.RegisterRequest,
) (*authv1.TokenPairResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.config.Business.ContextTimeout)
	defer cancel()

	return c.auth.Register(ctx, req)
}

func (c *Client) Login(
	ctx context.Context,
	req *authv1.LoginRequest,
) (*authv1.TokenPairResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.config.Business.ContextTimeout) //зачем мы передаем контекст? на что этот таймаут? зачем вообще нужны контексты?
	defer cancel()

	return c.auth.Login(ctx, req)
}

func (c *Client) Refresh(
	ctx context.Context,
	req *authv1.RefreshRequest,
) (*authv1.TokenPairResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.config.Business.ContextTimeout)
	defer cancel()

	return c.auth.Refresh(ctx, req)
}

func (c *Client) Logout(
	ctx context.Context,
	req *authv1.RefreshRequest,
) (*authv1.LogoutResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.config.Business.ContextTimeout)
	defer cancel()

	return c.auth.Logout(ctx, req)
}
