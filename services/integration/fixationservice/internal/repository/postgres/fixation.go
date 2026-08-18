package postgres

import (
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	Tx *TxManager
}

func NewRepository(tx *TxManager) *Repository {
	return &Repository{Tx: tx}
}

func (r *Repository) FixationCurrent(ctx context.Context, phoneHash string, projectID uuid.UUID) (*entity.Fixation, error) {
	const op = "postgres/FixationCurrent"
	f := &entity.Fixation{}
	query := `SELECT f.id, f.fixed_at, f.expires_at, f.status, f.agency_id, f.fix_by, f.fix_for, f.project_id, f.phone_hash FROM fixations f  
                WHERE f.phone_hash = $1 AND f.project_id = $2 
                                 ORDER BY f.fixed_at DESC LIMIT 1;`

	err := r.Tx.Querier(ctx).QueryRow(ctx, query, phoneHash, projectID).Scan(
		&f.FixationID,
		&f.FixedAt,
		&f.ExpiresAt,
		&f.Status,
		&f.AgencyID,
		&f.FixBy,
		&f.FixFor,
		&f.ProjectID,
		&f.PhoneHash,
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
	query := `SELECT 1 FROM app.projects p WHERE p.id = $1;`
	var count int
	err := r.Tx.Querier(ctx).QueryRow(ctx, query, projectID).Scan(&count)
	if err != nil {
		return false, MapError(op, err)
	}

	return true, nil
}

func (r *Repository) IsUserIDInAgencyID(ctx context.Context, agencyID, userID uuid.UUID) (bool, error) {
	const op = "postgres/IsUserIDInAgencyID"
	query := `SELECT 1 FROM app.users u WHERE u.agency_id = $1 AND u.user_id = $2;`
	var count int
	err := r.Tx.Querier(ctx).QueryRow(ctx, query, agencyID, userID).Scan(&count)
	if err != nil {
		return false, MapError(op, err)
	}

	return true, nil
}

func (r *Repository) InsertNewFixation(ctx context.Context, f entity.Fixation) error {
	const op = "postgres.fixation.InsertNewFixation"

	query := `INSERT INTO fixations(id, fixed_at, expires_at, status, agency_id, fixed_by, fix_for, project_id, phone_hash)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	_, err := r.Tx.Querier(ctx).Exec(ctx, query, f.FixationID, f.FixedAt, f.ExpiresAt, f.Status, f.AgencyID, f.FixBy, f.FixFor, f.ProjectID, f.PhoneHash)
	if err != nil {
		return MapError(op, err)
	}

	return nil
}

func (r *Repository) InsertAudit(ctx context.Context, f entity.Fixation) error {
	const op = "postgres.fixation.InsertAudit"

	query := `INSERT INTO audit(id, fixed_at, expires_at, status, agency_id, fixed_by, fix_for, project_id, phone_hash)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	_, err := r.Tx.Querier(ctx).Exec(ctx, query, f.FixationID, f.FixedAt, f.ExpiresAt, f.Status, f.AgencyID, f.FixBy, f.FixFor, f.ProjectID, f.PhoneHash)
	if err != nil {
		return MapError(op, err)
	}

	return nil
}

func (r *Repository) InsertOutbox(ctx context.Context, f entity.Fixation) error {
	const op = "postgres.fixation.InsertOutbox"

	query := `INSERT INTO outbox(id, fixed_at, expires_at, status, agency_id, fixed_by, fix_for, project_id, phone_hash) 
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	_, err := r.Tx.Querier(ctx).Exec(ctx, query, f.FixationID, f.FixedAt, f.ExpiresAt, f.Status, f.AgencyID, f.FixBy, f.FixFor, f.ProjectID, f.PhoneHash)
	if err != nil {
		return MapError(op, err)
	}

	return nil
}

func (r *Repository) UpdateFixationStatusExpired(ctx context.Context, statusExpired entity.Status, id uuid.UUID) error {
	const op = "postgres.UpdateFixationStatusExpired"
	query := `UPDATE fixations f SET f.status = $1 WHERE f.id = $2;`

	tag, err := r.Tx.Querier(ctx).Exec(ctx, query, statusExpired, id)
	if err != nil {
		return MapError(op, err)
	}
	if tag.RowsAffected() != 1 {
		return MapError(op, fmt.Errorf("fixation status has not update"))
	}
	return nil
}
