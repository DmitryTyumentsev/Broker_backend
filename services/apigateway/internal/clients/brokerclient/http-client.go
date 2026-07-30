package brokerclient

import (
	"Broker_backend/services/apigateway/internal/config"
	"Broker_backend/services/apigateway/internal/transport/http/dto/brokerdto"
	"context"
	"net/http"

	"github.com/google/uuid"
)

type HTTPClient struct {
	httpClient *http.Client
	cfg        *config.Config
}

func NewHTTPClient(cfg *config.Config) *HTTPClient {
	return &HTTPClient{
		httpClient: &http.Client{},
		cfg:        cfg,
	}
} //как установить соединение с brokerservice транспортом?

func (c *HTTPClient) NewFixation(ctx context.Context, req *brokerdto.FixationRequest, agencyID, userID uuid.UUID) (*brokerdto.FixationResponse, error) {
	return c.NewFixation(ctx, req, agencyID, userID)
}
