package postgres

import (
	"Donate_backend/services/authservice/internal/config"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Pool *pgxpool.Pool
	Cfg  *config.Config
}

func NewPostgres(pool *pgxpool.Pool, cfg *config.Config) *Postgres {
	return &Postgres{
		Pool: pool,
		Cfg:  cfg,
	}
}

func (pg *Postgres) WriteWithTimeout(ctx context.Context) context.Context {
	ctx, cancel := context.WithTimeout(ctx, pg.Cfg.Database.Postgres.WriteTimeout)
	defer cancel()
	return ctx
}

func (pg *Postgres) ReadWithTimeout(ctx context.Context) context.Context {
	ctx, cancel := context.WithTimeout(ctx, pg.Cfg.Database.Postgres.ReadTimeout)
	defer cancel()
	return ctx
}
