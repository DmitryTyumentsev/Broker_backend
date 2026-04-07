package postgres

import (
	"Donate_backend/services/authservice/internal/domain"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func MapError(err error, op string) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return domain.ErrNotUnique
		case pgerrcode.NotNullViolation:
			return domain.ErrMustBeNotNull
		case pgerrcode.TooManyRows:
			return domain.ErrBadRequest
		case pgerrcode.StringDataRightTruncationDataException:
			return domain.ErrBadRequest
		case pgerrcode.NoDataFound:
			return domain.ErrNotFound
		}
		return fmt.Errorf("op: %s, code: %v, err: %w", op, pgErr.Code, err)
	}
}
