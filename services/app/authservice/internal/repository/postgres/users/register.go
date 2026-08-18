package users

import (
	"Broker_backend/services/app/authservice/internal/domain/entity"
	"Broker_backend/services/app/authservice/internal/repository/postgres"
	"context"
)

type Repository struct {
	pg *postgres.Postgres
}

func NewRepository(pg *postgres.Postgres) *Repository {
	return &Repository{
		pg: pg,
	}
}

func (r *Repository) Save(ctx context.Context, user entity.User) error {
	const op = "postgres.users.Save"

	ctx, cancel := r.pg.WriteWithTimeout(ctx)
	defer cancel()

	query := `
		insert into app.users (
			id,
			agency_id,
			email,
			user_role,
			password_hash,
			last_name,
			first_name,
			middle_name,
			created_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pg.DB().Exec(
		ctx,
		query,
		user.ID,
		user.AgencyID,
		user.Email,
		user.Role,
		user.PasswordHash,
		user.LastName,
		user.FirstName,
		user.MiddleName,
		user.CreatedAt,
	)
	if err != nil {
		return postgres.MapError(op, err)
	}

	return nil
}
