package postgres

import (
	"Broker_backend/services/integration/fixationservice/internal/domain"
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"context"
	"time"
)

func (r *Repository) FetchUnpublished(ctx context.Context, limit int) ([]entity.Outbox, error) {
	const op = "postgres.FetchUnpublished"
	query := `select o.id, o.aggregate_id, o.aggregate_type, o.event_type, o.payload, o.created_at from integration.outbox o 
    where o.published_at is null order by o.id asc limit $1`

	rows, err := r.Tx.Querier(ctx).Query(ctx, query, limit)
	if err != nil {
		return nil, MapError(op, err)
	}
	defer rows.Close()

	events := make([]entity.Outbox, 0)

	for rows.Next() {
		outbox := &entity.Outbox{}

		err = rows.Scan(
			&outbox.ID,
			&outbox.ObjectID,
			&outbox.ObjectType,
			&outbox.EventType,
			&outbox.Payload,
			&outbox.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		events = append(events, *outbox)
	}
	err = rows.Err() // верно понял что после rows.Next() то есть после цикла rows.Err() ловит err из цикла и мапит их?
	if err != nil {
		return nil, MapError(op, err)
	}

	return events, nil
}

func (r *Repository) MarkPublished(ctx context.Context, ids []int, now time.Time) error {
	const op = "postgres.MarkPublished"
	query := `update integration.outbox o set values(o.published_at = $1)
    where id = $2`

	tag, err := r.Tx.Querier(ctx).Exec(ctx, query, now, ids)
	if err != nil {
		return MapError(op, err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
