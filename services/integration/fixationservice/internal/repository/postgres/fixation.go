package postgres

import (
	"Broker_backend/services/integration/fixationservice/internal/domain"
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
	const op = "postgres.fixation.FixationCurrent"
	f := &entity.Fixation{}
	query := `SELECT f.id, f.fixed_at, f.expires_at, f.status, f.agency_id,
       f.fix_by, f.fix_for, f.project_id, f.phone_hash
  FROM integration.fixations AS f
 WHERE f.phone_hash = $1 AND f.project_id = $2
 ORDER BY f.fixed_at DESC, f.id DESC
 LIMIT 1
 FOR UPDATE;`

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

func (r *Repository) StatusByProjectID(ctx context.Context, projectID uuid.UUID) (string, error) {
	const op = "postgres.fixation.StatusByProjectID"
	query := `SELECT p.status FROM app.projects p WHERE p.id = $1 ORDER BY p.id DESC LIMIT 1`
	var status string
	err := r.Tx.Querier(ctx).QueryRow(ctx, query, projectID).Scan(&status)
	if err != nil {
		return "", MapError(op, err)
	}

	return status, nil
}

func (r *Repository) IsUserIDInAgencyID(ctx context.Context, agencyID, userID uuid.UUID) (bool, error) {
	const op = "postgres.fixation.IsUserIDInAgencyID"
	query := `SELECT EXISTS(
    SELECT 1
      FROM app.users AS u
     WHERE u.agency_id = $1 AND u.id = $2
);`
	err := r.Tx.Querier(ctx).QueryRow(ctx, query, agencyID, userID).Scan()
	if err != nil {
		return false, MapError(op, err)
	}

	return true, nil
}

func (r *Repository) InsertNewFixation(ctx context.Context, f entity.Fixation) error {
	const op = "postgres.fixation.InsertNewFixation"

	query := `INSERT INTO integration.fixations(id, fixed_at, expires_at, status, agency_id, fix_by, fix_for, project_id, phone_hash)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	_, err := r.Tx.Querier(ctx).Exec(ctx, query, f.FixationID, f.FixedAt, f.ExpiresAt, f.Status, f.AgencyID, f.FixBy, f.FixFor, f.ProjectID, f.PhoneHash)
	if err != nil {
		return MapError(op, err)
	}

	return nil
}

func (r *Repository) InsertAudit(ctx context.Context, f entity.Fixation) error {
	const op = "postgres.fixation.InsertAudit"

	stateAfter, err := marshalFixationState(f)
	if err != nil {
		return fmt.Errorf("%s: marshal state_after: %w", op, err)
	}

	actor := auditActorFromContext(ctx)
	query := `INSERT INTO integration.audit_log(
    entity_type, entity_id, action, state_before, state_after,
    actor_user_id, actor_agency_id, actor_role, request_id, client_ip
) VALUES($1, $2, $3, NULL, $4::jsonb, $5, $6, $7, $8, $9::inet);`
	_, err = r.Tx.Querier(ctx).Exec(
		ctx,
		query,
		fixationAggregateType,
		f.FixationID,
		fixationCreatedAction,
		stateAfter,
		actor.UserID,
		actor.AgencyID,
		actor.Role,
		actor.RequestID,
		actor.ClientIP,
	)
	if err != nil {
		return MapError(op, err)
	}

	return nil
}

func (r *Repository) InsertOutbox(ctx context.Context, f entity.Fixation) error {
	const op = "postgres.fixation.InsertOutbox"

	payload, err := marshalFixationState(f)
	if err != nil {
		return fmt.Errorf("%s: marshal payload: %w", op, err)
	}

	query := `INSERT INTO integration.outbox(
    aggregate_type, aggregate_id, event_type, payload
) VALUES($1, $2, $3, $4::jsonb);`
	_, err = r.Tx.Querier(ctx).Exec(
		ctx,
		query,
		fixationAggregateType,
		f.FixationID,
		fixationCreatedEventType,
		payload,
	)
	if err != nil {
		return MapError(op, err)
	}

	return nil
}

func (r *Repository) UpdateFixationStatusExpired(ctx context.Context, statusExpired entity.Status, id uuid.UUID) error {
	const op = "postgres.fixation.UpdateFixationStatusExpired"
	if statusExpired != entity.StatusExpired {
		return fmt.Errorf("%s: expected status %q: %w", op, entity.StatusExpired, domain.ErrBadRequest)
	}

	query := `UPDATE integration.fixations
   SET status = $1
 WHERE id = $2 AND status = 'active';`

	tag, err := r.Tx.Querier(ctx).Exec(ctx, query, statusExpired, id)
	if err != nil {
		return MapError(op, err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("%s: active fixation %s: %w", op, id, domain.ErrNotFound)
	}

	return nil
}
