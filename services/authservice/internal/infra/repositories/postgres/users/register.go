package users

import (
	"Donate_backend/services/authservice/internal/domain/entity"
	postgres "Donate_backend/services/authservice/internal/infra/repositories/postgres"
	"context"
)

type Database struct {
	pg *postgres.Postgres
}

func NewDatabase(pg *postgres.Postgres) *Database {
	return &Database{
		pg: pg,
	}
}

func (db *Database) Save(ctx context.Context, user *entity.User) error {
	const op = "users.Save"
	query := `insert into users(email, password, username) VALUES ($1, $2, $3)`
	ctx = db.pg.WriteWithTimeout(ctx)
	_, err := db.pg.Pool.Exec(ctx, query, user.Email, user.PassHash, user.Username)
	if err != nil {
		return postgres.MapError(err, op)
	}

	return nil
}
