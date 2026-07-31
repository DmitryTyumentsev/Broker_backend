package postgres

import (
	"Broker_backend/services/integration/fixationservice/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
	cfg  *config.Config
}

func NewPostgres(db *pgxpool.Pool, cfg *config.Config) *Postgres {
	return &Postgres{
		pool: db,
		cfg:  cfg,
	}
}
