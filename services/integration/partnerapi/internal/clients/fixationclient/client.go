package fixationclient

import (
	fixationv1 "Broker_backend/gen/fixation/v1"
	"Broker_backend/services/integration/partnerapi/internal/config"
	grpcauth "Broker_backend/shared/pkg/grpc/auth"
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	fixation fixationv1.FixationServiceClient
	config   *config.Config
}

func NewFixationServiceClient(cfg *config.Config) (*grpc.ClientConn, fixationv1.FixationServiceClient, error) {
	if cfg == nil {
		return nil, nil, errors.New("config is nil")
	}

	if cfg.FixationGRPC.Address == "" {
		return nil, nil, errors.New("fixation_grpc.address is required")
	}

	conn, err := grpc.NewClient(
		"passthrough:///"+cfg.FixationGRPC.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create fixation service client: %w", err)
	}

	return conn, fixationv1.NewFixationServiceClient(conn), nil
}

func NewClient(fixation fixationv1.FixationServiceClient, cfg *config.Config) (*GRPCClient, error) {
	client := &GRPCClient{
		fixation: fixation,
		config:   cfg,
	}

	if err := client.Validate(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *GRPCClient) Validate() error {
	switch {
	case c == nil:
		return errors.New("client is nil")
	case c.fixation == nil:
		return errors.New("fixation grpc client is nil")
	case c.config == nil:
		return errors.New("config is nil")
	case c.config.OperationTimeout() <= 0:
		return errors.New("operation timeout must be positive")
	default:
		return nil
	}
}

func (c *GRPCClient) contextWithTimeout(parent context.Context) (context.Context,
	context.CancelFunc, error) {
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

func (c *GRPCClient) NewFixation(ctx context.Context, req *fixationv1.NewFixationRequest) (
	*fixationv1.NewFixationResponse, error) {
	ctx, cancel, err := c.contextWithTimeout(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	return c.fixation.NewFixation(ctx, req)
}
