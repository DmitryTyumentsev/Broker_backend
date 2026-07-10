package postgres

import (
	"errors"
	"fmt"

	"Broker_backend/services/brokerservice/internal/domain"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	constraintUniqueEmail            = "users_email_unique"
	constraintCheckUserRole          = "users_user_role_check"
	constraintUniqueRefreshTokenHash = "refresh_sessions_refresh_token_hash_unique"
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
			return fmt.Errorf("%s: %w: %v", op, domain.ErrMustBeNotNull, err)

		case pgerrcode.ForeignKeyViolation:
			return fmt.Errorf("%s: %w: %v", op, domain.ErrBadRequest, err)

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

func mapUniqueViolation(op string, constraintName string) error {
	switch constraintName {
	case constraintUniqueEmail:
		return fmt.Errorf("%s: %w", op, domain.ErrNotUniqueEmail)
	case constraintUniqueRefreshTokenHash:
		return fmt.Errorf("%s: %w", op, domain.ErrNotUniqueRefreshTokenHash)
	default:
		return fmt.Errorf("%s: %w", op, domain.ErrNotUnique)
	}
}

func mapCheckViolation(op string, constraintName string) error {
	switch constraintName {
	case constraintCheckUserRole:
		return fmt.Errorf("%s: %w", op, domain.ErrUserRoleInvalid)
	default:
		return fmt.Errorf("%s: %w", op, domain.ErrBadRequest)
	}
}
