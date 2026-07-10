package usecases

import (
	"Broker_backend/services/brokerservice/internal/domain/interfaces"
	"Broker_backend/services/brokerservice/internal/infra/repositories/postgres"
	"fmt"

	"Broker_backend/services/brokerservice/internal/config"

	"go.uber.org/zap"
)

type Service struct {
	config *config.Config
	logger *zap.Logger
	clock  interfaces.Clock
	pg     *postgres.Postgres
}

func NewService(
	cfg *config.Config,
	logger *zap.Logger,
	clock interfaces.Clock,
	pg *postgres.Postgres,
) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Service{
		config: cfg,
		logger: logger,
		clock:  clock,
		pg:     pg,
	}
}

func (s *Service) ensureDeps() error {
	switch {
	case s == nil:
		return fmt.Errorf("service is nil")
	case s.config == nil:
		return fmt.Errorf("config is nil")
	case s.clock == nil:
		return fmt.Errorf("clock is nil")
	case s.pg == nil:
		return fmt.Errorf("postgres is nil")
	default:
		return nil
	}
}
