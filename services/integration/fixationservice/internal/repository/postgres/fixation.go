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

func (r *Repository) FixationStatusCurrent(ctx context.Context, phoneHash string, projectID uuid.UUID) (entity.Status, error) {
	const op = "postgres/FixationStatusCurrent"
	var st entity.Status
	query := `SELECT f.status from f fixations 
                where f.phone_hash = $1 and f.project_id = $2 
                                 order by f.id desc limit 1`

	err := r.tx.pool.QueryRow(ctx, query, phoneHash, projectID).Scan(&st)
	if err != nil {
		return "", MapError(op, err)
	}
	return st, nil
}

func (r *Repository) IsExistsProjectID(ctx context.Context, projectID uuid.UUID) bool {
	query := `SELECT 1 FROM f fixations WHERE f.project_id = $1`
	var count int
	_ = r.tx.pool.QueryRow(ctx, query, projectID).Scan(&count)

	return count > 0
}

func (r *Repository) IsUserIDInAgencyID(ctx context.Context, agencyID, userID uuid.UUID) bool {
	query := `SELECT 1 FROM u app.users WHERE u.agency_id = $1 AND u.user_id = $2`
	var count int
	_ = r.tx.pool.QueryRow(ctx, query, agencyID, userID).Scan(&count)

	return count > 0
}

func (r *Repository) InsertFixation(ctx context.Context, f entity.Fixation) error {
	const op = "postgres.fixation.InsertFixation"

	query := `INSERT INTO fixations(id, fixed_at, expires_at, status, broker_id, fixed_by, fix_for, project_id, phone_hash) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.tx.Querier(ctx).Exec(ctx, query, f.FixationID, f.FixedAt, f.ExpiresAt, f.Status, f.AgencyID, f.FixedBy, f.FixFor, f.ProjectID, f.PhoneHash)
	if err != nil {
		return MapError(op, err)
	}

	return nil
}

func (r *Repository) InsertAudit(ctx context.Context, f entity.Fixation) error {
	const op = "postgres.fixation.InsertAudit"

	query := `INSERT INTO audit(id, fixed_at, expires_at, status, broker_id, fixed_by, fix_for, project_id, phone_hash) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.tx.Querier(ctx).Exec(ctx, query, f.FixationID, f.FixedAt, f.ExpiresAt, f.Status, f.AgencyID, f.FixedBy, f.FixFor, f.ProjectID, f.PhoneHash)
	if err != nil {
		return MapError(op, err)
	}

	return nil
}

func (r *Repository) InsertOutbox(ctx context.Context, f entity.Fixation) error {
	const op = "postgres.fixation.InsertOutbox"

	query := `INSERT INTO outbox(id, fixed_at, expires_at, status, broker_id, fixed_by, fix_for, project_id, phone_hash) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.tx.Querier(ctx).Exec(ctx, query, f.FixationID, f.FixedAt, f.ExpiresAt, f.Status, f.AgencyID, f.FixedBy, f.FixFor, f.ProjectID, f.PhoneHash)
	if err != nil {
		return MapError(op, err)
	}

	return nil
}

//
//func (r *Repository) Update(ctx context.Context, f entity.Fixation) error {
//	const op = "postgres.fixation.Update"
//
//	query1 := `UPDATE fixations SET status = $1 WHERE fixation_id = $2`
//	query2 := `INSERT INTO fixations(id, fixed_at, expires_at, status, broker_id, fixed_by, fix_for, project_id, phone_hash) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
//
//	return r.tx.Do(ctx, func(ctx) error {
//		tag, err := r.tx.Querier(ctx).Exec(ctx, query1, f.StatusExpired, f.FixationIDOld)
//		if tag.RowsAffected() != 1 {
//			return MapError(op, pgx.ErrNoRows)
//		}
//		if err != nil {
//			return MapError(op, err)
//		}
//
//		_, err = r.tx.Querier(ctx).Exec(ctx, query2, f.FixationID, f.FixedAt, f.ExpiresAt, f.StatusActive, f.AgencyID, f.FixedBy, f.FixFor, f.ProjectID, f.PhoneHash)
//		if err != nil {
//			return MapError(op, err)
//		}
//		return nil
//	})
//
//}
