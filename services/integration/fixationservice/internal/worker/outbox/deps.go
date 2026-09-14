package outbox

import (
	"Broker_backend/services/integration/fixationservice/internal/config"
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Repository interface {
	FetchUnpublished(ctx context.Context, limit int) ([]entity.Outbox, error)
}

type EventSender interface {
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type Clock interface {
	Now() time.Time
}

type Worker struct {
	cfg        *config.Config
	logger     *zap.Logger
	clock      Clock
	repository Repository
	tx         TxManager
}

func NewWorker(
	cfg *config.Config,
	logger *zap.Logger,
	clock Clock,
	repository Repository,
	tx TxManager,
) *Worker {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Worker{
		cfg:        cfg,
		logger:     logger,
		clock:      clock,
		repository: repository,
		tx:         tx,
	}
}

func (w *Worker) ensureDeps() error {
	switch {
	case w == nil:
		return fmt.Errorf("worker is nil")
	case w.cfg == nil:
		return fmt.Errorf("config is nil")
	case w.clock == nil:
		return fmt.Errorf("clock is nil")
	case w.repository == nil:
		return fmt.Errorf("postgres is nil")
	case w.tx == nil:
		return fmt.Errorf("tx is nil")
	default:
		return nil
	}
}
