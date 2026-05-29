package users

import (
	"Broker_backend/services/authservice/internal/domain/entity"
	"Broker_backend/services/authservice/internal/infra/repositories/postgres"
	"context"
)

func (r *Repository) FindByID(ctx context.Context, id string) (entity.User, error) {
	const op = "postgres.users.FindByID"

	ctx, cancel := r.pg.ReadWithTimeout(ctx)
	defer cancel()

	query := `
		select
			id,
			email,
			user_role,
			password_hash,
			last_name,
			first_name,
			middle_name,
			created_at,
			updated_at
		from users
		where id = $1
	`

	var user entity.User

	err := r.pg.DB().QueryRow(ctx, query, id).Scan(
		&user.ID,
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

func (r *Repository) FindByEmail(ctx context.Context, email string) (entity.User, error) {
	const op = "postgres.users.FindByEmail"

	ctx, cancel := r.pg.ReadWithTimeout(ctx)
	defer cancel()

	query := `
		select
			id,
			email,
			user_role,
			password_hash,
			last_name,
			first_name,
			middle_name,
			created_at,
			updated_at
		from users
		where email = $1
	`

	var user entity.User

	err := r.pg.DB().QueryRow(ctx, query, email).Scan(
		&user.ID,
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
