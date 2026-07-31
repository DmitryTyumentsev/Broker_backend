package postgres

import (
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	tx *TxManager
}

func NewRepository(tx *TxManager) *Repository {
	return &Repository{tx: tx}
}

func (r *Repository) FindActiveFixation(ctx context.Context, phoneHash string, projectID uuid.UUID) (entity.FixationCustomer, error) {
	var f entity.FixationCustomer

	query := `SELECT `
	return f, nil
}

func (r *Repository) Insert(ctx context.Context, fixationCustomer entity.FixationCustomer) error {
	const op = "postgres.fixation.Insert"

	query := `INSERT INTO fixations(id, fixed_at, expires_at, status, broker_id, fixed_by, fix_for, project_id, phone_hash) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	//верно понял что надо перед вставкой проверить что фиксации нет? как это правильно сделать - сделать метод check в юзкейсах отдельный который будет делать select или сделать в этом методе select и insert транзакцией?
	_, err := r.tx.Querier(ctx).Exec(ctx, query, fixationCustomer.FixationIDNew, fixationCustomer.FixedAt, fixationCustomer.ExpiresAt, fixationCustomer.StatusActive, fixationCustomer.AgencyID, fixationCustomer.FixedBy, fixationCustomer.FixFor, fixationCustomer.ProjectID, fixationCustomer.PhoneHash)
	if err != nil {
		return MapError(op, err)
	}

	return nil
}

func (r *Repository) Update(ctx context.Context, f entity.FixationCustomer) error {
	const op = "postgres.fixation.Update"

	query1 := `UPDATE fixations SET status = $1 WHERE fixation_id = $2`
	query2 := `INSERT INTO fixations(id, fixed_at, expires_at, status, broker_id, fixed_by, fix_for, project_id, phone_hash) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	return r.tx.Do(ctx, func(context.Context) error {
		tag, err := r.tx.Querier(ctx).Exec(ctx, query1, f.StatusExpired, f.FixationIDOld) //я вставил r и очень удивлен что сюда подставились методы. Это в продолжение вопроса про интерфейсы, вот очередное применение где я не понимаю что как сделали и что как получилось по шагам
		if tag.RowsAffected() != 1 {
			return MapError(op, pgx.ErrNoRows)
		}
		if err != nil {
			return MapError(op, err)
		}

		_, err = r.tx.Querier(ctx).Exec(ctx, query2, f.FixationIDNew, f.FixedAt, f.ExpiresAt, f.StatusActive, f.AgencyID, f.FixedBy, f.FixFor, f.ProjectID, f.PhoneHash)
		if err != nil {
			return MapError(op, err)
		}
		return nil
	})

}
