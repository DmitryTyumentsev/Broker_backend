package postgres

import (
	"errors"
	"fmt"

	"Donate_backend/services/authservice/internal/domain"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	constraintUniqueEmail            = "users_email_unique"
	constraintCheckUserRole          = "users_user_role_check" //для самих const лучше писать сокращенное название?
	constraintUniqueRefreshTokenHash = "refresh_sessions_refresh_token_hash_unique"
)

func MapError(op string, err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return fmt.Errorf("%s: %w", op, domain.ErrNotFound)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			switch pgErr.ConstraintName {
			case constraintUniqueEmail:
				return fmt.Errorf("%s: %w", op, domain.ErrNotUniqueEmail)
			case constraintUniqueRefreshTokenHash:
				return fmt.Errorf("%s: %w", op, domain.ErrNotUniqueRefreshTokenHash)
			}

		case pgerrcode.NotNullViolation:
			return fmt.Errorf("%s: %w: %v", op, domain.ErrMustBeNotNull, err)

		case pgerrcode.ForeignKeyViolation:
			return fmt.Errorf("%s: %w: %v", op, domain.ErrBadRequest, err)

		case pgerrcode.CheckViolation:
			switch pgErr.ConstraintName {
			case constraintCheckUserRole:
				return fmt.Errorf("%s: %w", op, domain.ErrUserRoleInvalid)
			}
		}
		return fmt.Errorf("%s: postgres error code=%s: %w", op, pgErr.Code, err)
	}
	return fmt.Errorf("%s: %w", op, err)
}
