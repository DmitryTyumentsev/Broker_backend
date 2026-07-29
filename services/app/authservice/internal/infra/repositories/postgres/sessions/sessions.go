package sessions

import (
	"Broker_backend/services/app/authservice/internal/domain"
	"Broker_backend/services/app/authservice/internal/domain/entity"
	postgres2 "Broker_backend/services/app/authservice/internal/infra/repositories/postgres"
	"context"
	"fmt"
)

type Repository struct {
	pg *postgres2.Postgres
}

func NewRepository(pg *postgres2.Postgres) *Repository {
	return &Repository{
		pg: pg,
	}
}

func (r *Repository) Save(ctx context.Context, session entity.RefreshSession) error {
	const op = "postgres.sessions.Save"

	ctx, cancel := r.pg.WriteWithTimeout(ctx)
	defer cancel()

	query := `
		insert into refresh_sessions (
			refresh_token_hash,
			user_id,
			device_id,
			created_at,
			expires_at,
			revoked_at,
			replaced_by_refresh_token_hash
		)
		values ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pg.DB().Exec(
		ctx,
		query,
		session.RefreshTokenHash,
		session.UserID,
		session.DeviceID,
		session.CreatedAt,
		session.ExpiresAt,
		session.RevokedAt,
		session.ReplacedByRefreshTokenHash,
	)
	if err != nil {
		return postgres2.MapError(op, err)
	}

	return nil
}

func (r *Repository) FindByHash(ctx context.Context, hash string) (entity.RefreshSession, error) {
	const op = "postgres.sessions.FindByHash"

	ctx, cancel := r.pg.ReadWithTimeout(ctx)
	defer cancel()

	query := `
		select
			session_id,
			refresh_token_hash,
			user_id,
			device_id,
			created_at,
			expires_at,
			revoked_at,
			replaced_by_refresh_token_hash
		from refresh_sessions
		where refresh_token_hash = $1
	`

	var session entity.RefreshSession

	err := r.pg.DB().QueryRow(ctx, query, hash).Scan(
		&session.SessionID,
		&session.RefreshTokenHash,
		&session.UserID,
		&session.DeviceID,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.ReplacedByRefreshTokenHash,
	)
	if err != nil {
		return entity.RefreshSession{}, postgres2.MapError(op, err)
	}

	return session, nil
}

func (r *Repository) Rotate(
	ctx context.Context,
	oldHash string,
	newSession entity.RefreshSession,
) error {
	const op = "postgres.sessions.Rotate"

	ctx, cancel := r.pg.WriteWithTimeout(ctx)
	defer cancel()

	tx, err := r.pg.DB().Begin(ctx)
	if err != nil {
		return postgres2.MapError(op, err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	updateQuery := `
		update refresh_sessions
		set
			revoked_at = $1,
			replaced_by_refresh_token_hash = $2
		where refresh_token_hash = $3
		  and revoked_at is null
	`

	tag, err := tx.Exec(
		ctx,
		updateQuery,
		newSession.CreatedAt,
		newSession.RefreshTokenHash,
		oldHash,
	)
	if err != nil {
		return postgres2.MapError(op, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrNotFound)
	}

	insertQuery := `
		insert into refresh_sessions (
			refresh_token_hash,
			user_id,
			device_id,
			created_at,
			expires_at,
			revoked_at,
			replaced_by_refresh_token_hash
			)
		values ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = tx.Exec(
		ctx,
		insertQuery,
		newSession.RefreshTokenHash,
		newSession.UserID,
		newSession.DeviceID,
		newSession.CreatedAt,
		newSession.ExpiresAt,
		newSession.RevokedAt,
		newSession.ReplacedByRefreshTokenHash,
	)
	if err != nil {
		return postgres2.MapError(op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return postgres2.MapError(op, err)
	}

	return nil
}

func (r *Repository) Revoke(ctx context.Context, hash string) error {
	const op = "postgres.sessions.Revoke"

	ctx, cancel := r.pg.WriteWithTimeout(ctx)
	defer cancel()

	query := `
		update refresh_sessions
		set revoked_at = coalesce(revoked_at, now())
		where refresh_token_hash = $1
	`

	tag, err := r.pg.DB().Exec(ctx, query, hash)
	if err != nil {
		return postgres2.MapError(op, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, domain.ErrNotFound)
	}

	return nil
}
