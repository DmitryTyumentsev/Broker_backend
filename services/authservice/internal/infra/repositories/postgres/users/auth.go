package users

import (
	"Donate_backend/services/authservice/internal/domain/entity"
	"Donate_backend/services/authservice/internal/infra/repositories/postgres"
	"context"
)

func (r *Repository) FindByEmail(ctx context.Context, email string) (entity.User, error) {
	const op = "postgres.users.FindByEmail"

	ctx, cancel := r.pg.ReadWithTimeout(ctx)
	defer cancel()

	query := `
		select
			id,
			email,
			password_hash,
			username,
			role,
			created_at
		from users
		where email = $1
	`

	var user entity.User

	err := r.pg.DB().QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Username,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return entity.User{}, postgres.MapError(op, err)
	}

	return user, nil
}
