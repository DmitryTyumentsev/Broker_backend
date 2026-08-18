package postgres

import (
	"Broker_backend/services/integration/fixationservice/internal/domain"
	"errors"
	"fmt"

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
			return mapUniqueViolation(op, pgErr.ConstraintName)

		case pgerrcode.NotNullViolation:
			return fmt.Errorf("%s: %w: %s", op, domain.ErrMustBeNotNull, pgErr.Message)

		case pgerrcode.ForeignKeyViolation:
			return fmt.Errorf("%s: %w: %s", op, domain.ErrBadRequest, pgErr.Message)

		case pgerrcode.CheckViolation:
			return mapCheckViolation(op, pgErr.ConstraintName)
		}

		return fmt.Errorf(
			"%s: postgres error code=%s constraint=%s: %w",
			op,
			pgErr.Code,
			pgErr.ConstraintName,
			err,
		)
	}

	return fmt.Errorf("%s: %w", op, err)
}

// mapUniqueViolation пока не разбирает имя ограничения: доменная ошибка одна.
// Когда уникальных ограничений станет больше одного, здесь появится switch по
// constraintName — как сделано в authservice/repository/postgres/errors.go.
func mapUniqueViolation(op string, _ string) error {
	return fmt.Errorf("%s: %w", op, domain.ErrNotUnique)
}

// mapCheckViolation — то же самое для check-ограничений.
func mapCheckViolation(op string, _ string) error {
	return fmt.Errorf("%s: %w", op, domain.ErrBadRequest)
}
