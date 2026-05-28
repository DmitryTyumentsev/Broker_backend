package users

import (
	"Broker_backend/services/authservice/internal/domain/entity"
	"Broker_backend/services/authservice/internal/infra/repositories/postgres"
	"context"
)

func (r *Repository) FindByEmail(ctx context.Context, email string) (entity.User, error) {
	const op = "postgres.users.FindByEmail"

	ctx, cancel := r.pg.ReadWithTimeout(ctx)
	defer cancel()

	query := `select id, email, user_role, password_hash, last_name, first_name, middle_name, created_at
     from users where email = $1`

	var user entity.User

	err := r.pg.DB().QueryRow(ctx, query, email).Scan( //немного не понял почему правильно создавать метод DB(). это чтобы структура repository этого пакета могла подставляться вместо postgres.Postgres? не проще ли не реализуя через метод вызывать? немного не понял зачем так сделали, разве не только для интерфейсов так надо заморачиваться чтобы не ловить ошибку что не локальное вызвано? и почему в целом задумано что ошибку ловим если структуру с другого пакета ставим в зависимость в ресивер? в чем профит, зачем так сделано в языке?
		&user.ID,
		&user.Email,
		&user.Role,
		&user.PasswordHash,
		&user.LastName,
		&user.FirstName,
		&user.MiddleName,
		&user.CreatedAt,
	)
	if err != nil {
		return entity.User{}, postgres.MapError(op, err)
	}

	return user, nil
}
