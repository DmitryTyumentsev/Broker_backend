package postgres

import (
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	tx *TxManager
}

func NewRepository(tx *TxManager) *Repository {
	return &Repository{tx: tx}
}

func (r *Repository) FixationCurrent(ctx context.Context, phoneHash string, projectID uuid.UUID) (*entity.Fixation, error) {
	const op = "postgres/FixationCurrent"
	f := &entity.Fixation{}
	query := `SELECT f.id, f.fixed_at, f.expires_at, f.status, f.agency_id, f.fixed_by, f.fix_for, f.project_id, f.phone_hash FROM fixations f  
                WHERE f.phone_hash = $1 AND f.project_id = $2 
                                 ORDER BY f.id DESC LIMIT 1;`

	err := r.tx.Querier(ctx).QueryRow(ctx, query, phoneHash, projectID).Scan(
		f.FixationID,
		f.FixedAt,
		f.ExpiresAt,
		f.Status,
		f.AgencyID,
		f.FixedBy,
		f.FixFor,
		f.ProjectID,
		f.PhoneHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			f.Status = entity.StatusNoRows
			return f, nil
		}
		return nil, MapError(op, err)
	}
	return f, nil
}

func (r *Repository) IsExistsProjectID(ctx context.Context, projectID uuid.UUID) (bool, error) {
	const op = "postgres/IsExistsProjectID"
	query := `SELECT 1 FROM projects p WHERE p.id = $1;`
	var count int
	err := r.tx.Querier(ctx).QueryRow(ctx, query, projectID).Scan(&count)
	if err != nil {
		return false, MapError(op, err)
	}

	return true, nil
}

func (r *Repository) IsUserIDInAgencyID(ctx context.Context, agencyID, userID uuid.UUID) (bool, error) {
	const op = "postgres/IsUserIDInAgencyID"
	query := `SELECT 1 FROM app.users u WHERE u.agency_id = $1 AND u.user_id = $2;`
	var count int
	err := r.tx.Querier(ctx).QueryRow(ctx, query, agencyID, userID).Scan(&count)
	if err != nil {
		return false, MapError(op, err)
	}

	return true, nil
}

func (r *Repository) InsertNewFixation(ctx context.Context, f entity.Fixation) error {
	const op = "postgres.fixation.InsertNewFixation"

	query := `INSERT INTO fixations(id, fixed_at, expires_at, status, agency_id, fixed_by, fix_for, project_id, phone_hash)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	_, err := r.tx.Querier(ctx).Exec(ctx, query, f.FixationID, f.FixedAt, f.ExpiresAt, f.Status, f.AgencyID, f.FixedBy, f.FixFor, f.ProjectID, f.PhoneHash)
	if err != nil {
		return MapError(op, err)
	}

	return nil
}

func (r *Repository) InsertAudit(ctx context.Context, f entity.Fixation) error {
	const op = "postgres.fixation.InsertAudit"

	query := `INSERT INTO audit(id, fixed_at, expires_at, status, agency_id, fixed_by, fix_for, project_id, phone_hash)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	_, err := r.tx.Querier(ctx).Exec(ctx, query, f.FixationID, f.FixedAt, f.ExpiresAt, f.Status, f.AgencyID, f.FixedBy, f.FixFor, f.ProjectID, f.PhoneHash)
	if err != nil {
		return MapError(op, err)
	}

	return nil
}

func (r *Repository) InsertOutbox(ctx context.Context, f entity.Fixation) error {
	const op = "postgres.fixation.InsertOutbox"

	query := `INSERT INTO outbox(id, fixed_at, expires_at, status, agency_id, fixed_by, fix_for, project_id, phone_hash) 
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	_, err := r.tx.Querier(ctx).Exec(ctx, query, f.FixationID, f.FixedAt, f.ExpiresAt, f.Status, f.AgencyID, f.FixedBy, f.FixFor, f.ProjectID, f.PhoneHash)
	if err != nil {
		return MapError(op, err)
	}

	return nil
}

func (r *Repository) UpdateFixationStatusRemoved(ctx context.Context, statusRemoved entity.Status, id uuid.UUID) error {
	const op = "postgres.UpdateFixationStatusRemoved"
	query := `with = SELECT f.id FROM fixations f WHERE f.id = $2 ORDER BY f.id DESC LIMIT 1;
UPDATE fixations f SET f.status = $1 WHERE f.id = with;`

	tag, err := r.tx.Querier(ctx).Exec(ctx, query, statusRemoved, id)
	if tag.RowsAffected() != 1 {
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
