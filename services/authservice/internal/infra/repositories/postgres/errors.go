package postgres

import (
	"errors"
	"fmt"

	"Donate_backend/services/authservice/internal/domain"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func MapError(op string, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w", op, domain.ErrNotFound)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return fmt.Errorf("%s: %w", op, domain.ErrNotUnique)
		case pgerrcode.NotNullViolation:
			return fmt.Errorf("%s: %w", op, domain.ErrMustBeNotNull)
		case pgerrcode.ForeignKeyViolation:
			return fmt.Errorf("%s: %w", op, domain.ErrBadRequest)
		default:
			return fmt.Errorf("%s: postgres error code=%s: %w", op, pgErr.Code, err)
		}
	}

	return fmt.Errorf("%s: %w", op, err)
}
