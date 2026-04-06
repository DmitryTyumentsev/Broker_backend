package users

import (
	"Donate_backend/services/authservice/internal/domain/entity"
	"Donate_backend/services/authservice/internal/infra/repositories/postgres"
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
	query := `insert into users(email, password_hash, username, created_at) VALUES ($1, $2, $3, $4) on coflict do nothing`
	ctx = db.pg.WriteWithTimeout(ctx)

	_, err := db.pg.Pool.Exec(ctx, query, user.Email, user.PassHash, user.Username, user.CreatedAt)
	if err != nil {
		return postgres.MapError(err, op)
	}

	return nil
}

func (db *Database) GetUserByID(ctx context.Context, uuid string) (*entity.User, error) {
	const op = "users.GetUserByID"
	query := `select * from users where uuid = $1 returnning email, password_hash, username, created_at`
	ctx = db.pg.ReadWithTimeout(ctx)
	var user *entity.User

	err := db.pg.Pool.QueryRow(ctx, query, uuid).Scan(&user)
	if err != nil {
		return nil, postgres.MapError(err, op)
	}

	return user, nil
}

func (db *Database) UpdateUsername(ctx context.Context, oldUsername, newUsername string) error {
	const op = "users.UpdateUsername"
	query := `update users set username = $1 where username = $2`
	ctx = db.pg.WriteWithTimeout(ctx)

	_, err := db.pg.Pool.Exec(ctx, query, newUsername, oldUsername)
	if err != nil {
		return postgres.MapError(err, op)
	}

	return nil
}

func (db *Database) DeleteUserByID(ctx context.Context, uuid string) error {
	const op = "users.DeleteUserByID"
	query := `delete * from users where uuid = $1`
	ctx = db.pg.WriteWithTimeout(ctx)

	_, err := db.pg.Pool.Exec(ctx, query, uuid)
	if err != nil {
		return postgres.MapError(err, op)
	}

	return nil
}
