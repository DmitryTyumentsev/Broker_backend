package users

import (
	"Donate_backend/services/authservice/internal/domain/entity"
	"Donate_backend/services/authservice/internal/repositories/postgres"
	"context"
	"errors"
	"fmt"
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
	//TODO: обязательно ли писать капсом запрос? и второй вопрос - чаще пользуются pg4admin,
	//dbeaver или cmd чтобы посмотреть что там в базе по итогу? и третий вопрос - в базах есть инкременты.
	err := db.pg.Pool.WriteWithTimeout(ctx, query)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, pgErr) {
			switch pgErr.Code {
			case pgerror.NotNullViolation: //TODO: поправить пакет
				return postgres.ErrNotNullViolation
			case pgerror.ForeignKeyViolation:
				return postgres.ErrForeignKeyViolation
			case pgerror.ErrInvalidViolation:
				return postgres.ErrInvalidViolation
			case pgerror.ErrTypeViolation:
				return postgres.ErrTypeViolation
			default:
				return fmt.Errorf("pgxerror, err: %w, query: %s", err, query)
			}
		}
		return fmt.Errorf("postgres error, err: %w, query: %s", err, query)
	}
	return err //TODO: лучше ли в return просто указать err или лучше когда большой кусок кода в return?
}
