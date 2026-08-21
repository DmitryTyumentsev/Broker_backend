package usecase

import (
	"Broker_backend/services/integration/fixationservice/internal/config"
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type FixationRepository interface {
	StatusByProjectID(ctx context.Context, projectID uuid.UUID) (string, error)
	IsUserIDInAgencyID(ctx context.Context, agencyID, userID uuid.UUID) (bool, error)
	FixationCurrent(ctx context.Context, phoneHash string, projectID uuid.UUID) (*entity.Fixation, error)
	InsertNewFixation(ctx context.Context, f entity.Fixation) error
	InsertAudit(ctx context.Context, f entity.Fixation) error
	InsertOutbox(ctx context.Context, f entity.Fixation) error
	UpdateFixationStatusExpired(ctx context.Context, statusExpired entity.Status, id uuid.UUID) error
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type Clock interface {
	Now() time.Time
}

type Service struct {
	cfg       *config.Config
	logger    *zap.Logger
	clock     Clock
	fixations FixationRepository
	tx        TxManager
}

func NewService(
	cfg *config.Config,
	logger *zap.Logger,
	clock Clock,
	fixation FixationRepository,
	tx TxManager,
) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Service{
		cfg:       cfg,
		logger:    logger,
		clock:     clock,
		fixations: fixation,
		tx:        tx,
	}
}

func (s *Service) ensureDeps() error {
	switch {
	case s == nil:
		return fmt.Errorf("service is nil")
	case s.cfg == nil:
		return fmt.Errorf("config is nil")
	case s.clock == nil:
		return fmt.Errorf("clock is nil")
	case s.fixations == nil:
		return fmt.Errorf("postgres is nil")
	case s.tx == nil:
		return fmt.Errorf("tx is nil")
	default:
		return nil
	}
}
