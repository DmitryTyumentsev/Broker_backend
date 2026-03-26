package users

import (
	"Donate_backend/services/authservice/internal/domain/entity"
	"Donate_backend/services/authservice/internal/repositories/postgres"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type Database struct {
	pg *postgres.Postgres
}

func NewDatabase(pg *postgres.Postgres) *Database {
	return &Database{
		pg: pg,
	}
}

func (db *Database) Save(ctx context.Context, user entity.User) error {
	const op = "users.Save()"
	query := `insert into users(email, pass, username) VALUES ($1, $2, $3)`
	err := db.pg.WriteWithTimeout(ctx, query)
	if err != nil {
		pgErr := new(pgconn.PgError)
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.NotNullViolation: //TODO: поправить пакет
				return postgres.ErrNotNullViolation
			case pgerrcode.ForeignKeyViolation:
				return postgres.ErrForeignKeyViolation
			case pgerrcode.ErrInvalidViolation:
				return postgres.ErrInvalidViolation
			case pgerrcode.ErrTypeViolation:
				return postgres.ErrTypeViolation
			default:
				return fmt.Errorf("pgxerror, err: %w, query: %s", err, query)
			}
		}
		return fmt.Errorf("postgres error, err: %w, query: %s", err, query)
	}
	return err //TODO: лучше ли в return просто указать err или лучше когда большой кусок кода в return?
}
