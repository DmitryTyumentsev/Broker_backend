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
	constraintUniqueEmail            = "unique_users_email"
	constraintCheckUserRole          = "check_users_user_role" //для самих const лучше писать сокращенное название?
	constraintUniqueRefreshTokenHash = "unique_refresh_sessions_refresh_token_hash"
)

func MapError(op string, err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return fmt.Errorf("%s: %w", op, domain.ErrNotFound) //верно ли это тут оставлять или можно ниже в case добавить?
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			switch pgErr.ConstraintName {
			case constraintUniqueEmail:
				return fmt.Errorf("%s: %w", op, domain.ErrNotUniqueEmail)
			case constraintUniqueRefreshTokenHash:
				return fmt.Errorf("%s: %w", op, domain.ErrNotUniqueRefreshTokenHash) //верно рассуждаю что надо возвращать ошибку в доменный слой чтобы сделать новую попытку создания хэша?
			} //мне немного не нравится switch в switch, ок ли так или можно более читаемо сделать? как на продовых проектах делают? или как щас у меня хорошо читаемо? и еще вопрос - зачем нужен pgerrcode.IntegrityConstraintViolation?

		case pgerrcode.NotNullViolation:
			return fmt.Errorf("%s: %w, full error: %w", op, domain.ErrMustBeNotNull, err)

		case pgerrcode.ForeignKeyViolation:
			return fmt.Errorf("%s: %w, full error: %w", op, domain.ErrBadRequest, err) //внешне получился винергрет внутри самого fmt.Errorf как по мне. Как тебе? или норм?

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
