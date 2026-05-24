package postgres

import (
	"context"

	"Donate_backend/services/authservice/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewPostgres(db *pgxpool.Pool, cfg *config.Config) *Postgres {
	return &Postgres{
		db:  db,
		cfg: cfg,
	}
}

func (p *Postgres) DB() *pgxpool.Pool {
	return p.db
}

func (p *Postgres) WriteWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, p.cfg.Database.Postgres.WriteTimeout)
}

func (p *Postgres) ReadWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, p.cfg.Database.Postgres.ReadTimeout)
}
