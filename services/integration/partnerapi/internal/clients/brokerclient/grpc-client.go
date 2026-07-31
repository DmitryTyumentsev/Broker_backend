package brokerclient

import (
	"Broker_backend/services/integration/partnerapi/internal/config"
	grpcauth "Broker_backend/shared/pkg/grpc/auth"
	brokerv1 "Broker_backend/shared/pkg/grpc/gen/broker/v1"
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// эталон ли больших проектов мой текущий auth-client.go? что мне нужно знать оттуда чтобы быть готовым к собесам на миддл+ ? спрашивают ли вообще это?
type GRPCClient struct {
	broker brokerv1.BrokerServiceClient
	config *config.Config
}

func NewBrokerServiceClient(cfg *config.Config) (*grpc.ClientConn, brokerv1.BrokerServiceClient, error) {
	if cfg == nil {
		return nil, nil, errors.New("config is nil")
	}

	if cfg.BrokerGRPC.Address == "" {
		return nil, nil, errors.New("broker_grpc.address is required")
	}

	conn, err := grpc.NewClient(
		"passthrough:///"+cfg.BrokerGRPC.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create broker service client: %w", err)
	}

	return conn, brokerv1.NewBrokerServiceClient(conn), nil
}

func NewClient(broker brokerv1.BrokerServiceClient, cfg *config.Config) (*GRPCClient, error) {
	client := &GRPCClient{
		broker: broker,
		config: cfg,
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
	case c.broker == nil:
		return errors.New("broker grpc client is nil")
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
	ctx = grpcauth.InjectOutgoingContext(ctx) //поправить и там внутри тоже
	return ctx, cancel, nil
}

func (c *GRPCClient) NewFixationCustomer(ctx context.Context, req *brokerv1.NewFixationCustomerRequest) (
	*brokerv1.NewFixationCustomerResponse, error) {
	ctx, cancel, err := c.contextWithTimeout(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	return c.broker.NewFixationCustomer(ctx, req)
}
