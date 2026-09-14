package outbox

import (
	"context"
)

const (
	limit = 100
)

func (w *Worker) Run(ctx context.Context) error {
	batch, err := w.repository.FetchUnpublished(ctx, limit)
	if err != nil {
		return err
	}

	return nil
}
