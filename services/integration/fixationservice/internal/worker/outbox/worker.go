package outbox

import (
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
)

const (
	batchUnpublished = 100
)

type EventSender interface {
	Send(ctx context.Context, ev []entity.Outbox) error
}
type Repository interface {
	FetchUnpublished(ctx context.Context, limit int) ([]entity.Outbox, error)
	MarkPublished(ctx context.Context, ids []int, now time.Time) error
}
type Config struct {
	Enabled   bool
	Tick      time.Duration
	BatchSize int
}
type Worker struct {
	repo   Repository
	sender EventSender
	log    *zap.Logger
	cfg    Config
}

func NewWorker(repo Repository, sender EventSender, log *zap.Logger, cfg Config) (*Worker, error) {
	switch {
	case repo == nil:
		return nil, errors.New("outbox: repo is nil")
	case sender == nil:
		return nil, errors.New("outbox: sender is nil")
	case cfg.Tick <= 0:
		return nil, errors.New("outbox: tick must be positive")
	case cfg.BatchSize <= 0:
		return nil, errors.New("outbox: batch_size must be positive")
	}
	if log == nil {
		log = zap.NewNop()
	}
	return &Worker{repo: repo, sender: sender, log: log, cfg: cfg}, nil
}
func (w *Worker) Run(ctx context.Context) error {
	t := time.NewTicker(w.cfg.Tick)
	defer t.Stop() // незакрытый тикер в долгоживущем
	// процессе — утечка таймера
	w.log.Info("outbox worker started",
		zap.Duration("tick", w.cfg.Tick), zap.Int("batch", w.cfg.BatchSize))
	for {
		select {
		case <-ctx.Done(): // ЕДИНСТВЕННЫЙ выход. Убить
			w.log.Info("outbox worker stopped") // горутину снаружи нельзя,
			return nil                          // её можно только попросить
		case <-t.C:
			// Полная пачка означает «в очереди есть ещё». Ждать тикер
			// в этом случае бессмысленно: после двухчасового простоя
			// CRM разгребание пойдёт со скоростью «пачка за тик»,
			// и десять тысяч событий будут ехать час
			for {
				n, err := w.processBatch(ctx)
				if err != nil {
					w.log.Error("outbox batch failed", zap.Error(err))
					break // ошибка пачки НЕ убивает воркер:
				} // плохие данные — это данные, а не авария
				if n < w.cfg.BatchSize || ctx.Err() != nil {
					break // неполная пачка → ждём тикер
				}
			}
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) (int, error) {
	b, err := w.repo.FetchUnpublished(ctx, batchUnpublished)
	if err != nil {
		return 0, err
	}
	err = w.sender.Send(ctx, b)
	if err != nil {
		return len(b), err
	}

	ids := make([]int, 0, len(b))

	for _, ev := range b {
		ids = append(ids, ev.ID)
	}
	now := time.Now().UTC()
	err = w.repo.MarkPublished(ctx, ids, now)
	if err != nil {
		return len(b), err
	}

	return len(b), nil
}
