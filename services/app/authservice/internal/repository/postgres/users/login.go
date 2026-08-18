package users

import (
	"Broker_backend/services/app/authservice/internal/domain/entity"
	"Broker_backend/services/app/authservice/internal/repository/postgres"
	"context"
)

// Схема указана явно. search_path в DSN настроен на app,public, но
// полагаться на него в запросах опасно: у мигратора и у сервиса он может
// разойтись, и запрос молча уедет не в ту схему.
const selectUserColumns = `
	select
		id,
		agency_id,
		email,
		user_role,
		password_hash,
		last_name,
		first_name,
		middle_name,
		created_at,
		updated_at
	from app.users
`

func (r *Repository) FindByID(ctx context.Context, id string) (entity.User, error) {
	const op = "postgres.users.FindByID"

	ctx, cancel := r.pg.ReadWithTimeout(ctx)
	defer cancel()

	return scanUser(op, r.pg.DB().QueryRow(ctx, selectUserColumns+` where id = $1`, id))
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (entity.User, error) {
	const op = "postgres.users.FindByEmail"

	ctx, cancel := r.pg.ReadWithTimeout(ctx)
	defer cancel()

	return scanUser(op, r.pg.DB().QueryRow(ctx, selectUserColumns+` where email = $1`, email))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(op string, row rowScanner) (entity.User, error) {
	var user entity.User

	err := row.Scan(
		&user.ID,
		&user.AgencyID,
		&user.Email,
		&user.Role,
		&user.PasswordHash,
		&user.LastName,
		&user.FirstName,
		&user.MiddleName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return entity.User{}, postgres.MapError(op, err)
	}

	return user, nil
}
