package users

import (
	"context"

	"Donate_backend/services/authservice/internal/domain/entity"
	"Donate_backend/services/authservice/internal/infra/repositories/postgres"
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
		insert into users (
		                   id,
		                   email,
		                   user_role,
		                   password_hash,
		                   last_name,
		                   first_name,
		                   middle_name,
		                   created_at,
		                   updated_at
		                   ) values ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pg.DB().Exec(
		ctx,
		query,
		user.ID,
		user.Email,
		user.Role,
		user.PassHash,
		user.LastName,
		user.FirstName, //выносят ли на реальных проектах такие значения как отдельно тип? чтобы не перепутать. насколько принято так писать в продовых проектах и если принято то где пишут?
		user.MiddleName,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return postgres.MapError(op, err)
	}

	return nil
}

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
		&user.PassHash,
		&user.Username,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return entity.User{}, postgres.MapError(op, err)
	}

	return user, nil
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (entity.User, error) {
	const op = "postgres.users.FindByUsername"

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
		where username = $1
	`

	var user entity.User

	err := r.pg.DB().QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.PassHash,
		&user.Username,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		return entity.User{}, postgres.MapError(op, err)
	}

	return user, nil
}
