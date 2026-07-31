package authclient

import (
	"Broker_backend/services/integration/partnerapi/internal/config"
	"context"
	"errors"
	"fmt"

	grpcauth "Broker_backend/shared/pkg/grpc/auth"
	authv1 "Broker_backend/shared/pkg/grpc/gen/auth/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	auth   authv1.AuthServiceClient
	config *config.Config
}

func NewAuthServiceClient(cfg *config.Config) (*grpc.ClientConn, authv1.AuthServiceClient, error) {
	if cfg == nil {
		return nil, nil, errors.New("config is nil")
	}

	if cfg.AuthGRPC.Address == "" {
		return nil, nil, errors.New("auth_grpc.address is required")
	}

	conn, err := grpc.NewClient(
		"passthrough:///"+cfg.AuthGRPC.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create auth service client: %w", err)
	}

	return conn, authv1.NewAuthServiceClient(conn), nil
}

func NewClient(auth authv1.AuthServiceClient, cfg *config.Config) (*Client, error) {
	client := &Client{
		auth:   auth,
		config: cfg,
	}

	if err := client.Validate(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) Validate() error {
	switch {
	case c == nil:
		return errors.New("client is nil")
	case c.auth == nil:
		return errors.New("auth grpc client is nil")
	case c.config == nil:
		return errors.New("config is nil")
	case c.config.OperationTimeout() <= 0:
		return errors.New("operation timeout must be positive")
	default:
		return nil
	}
}

func (c *Client) Register(
	ctx context.Context,
	req *authv1.RegisterRequest,
) (*authv1.RegisterResponse, error) {
	ctx, cancel, err := c.contextWithTimeout(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()

	return c.auth.Register(ctx, req)
}

func (c *Client) Login(
	ctx context.Context,
	req *authv1.LoginRequest,
) (*authv1.LoginResponse, error) {
	ctx, cancel, err := c.contextWithTimeout(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()

	return c.auth.Login(ctx, req)
}

func (c *Client) Refresh(
	ctx context.Context,
	req *authv1.RefreshRequest,
) (*authv1.RefreshResponse, error) {
	ctx, cancel, err := c.contextWithTimeout(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()

	return c.auth.Refresh(ctx, req)
}

func (c *Client) Logout(
	ctx context.Context,
	req *authv1.LogoutRequest,
) (*authv1.LogoutResponse, error) {
	ctx, cancel, err := c.contextWithTimeout(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()

	return c.auth.Logout(ctx, req)
}

func (c *Client) contextWithTimeout(parent context.Context) (context.Context, context.CancelFunc, error) {
	if err := c.Validate(); err != nil {
		return nil, nil, err
	}

	if parent == nil {
		parent = context.Background()
	}

	ctx, cancel := context.WithTimeout(parent, c.config.OperationTimeout())
	ctx = grpcauth.InjectOutgoingContext(ctx)

	return ctx, cancel, nil
}
