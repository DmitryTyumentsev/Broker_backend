package postgres

import (
	"Broker_backend/services/integration/brokerservice/internal/config"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	pgCfg := cfg.Database.Postgres

	poolCfg, err := pgxpool.ParseConfig(pgCfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	if pgCfg.MaxConnections > 0 {
		poolCfg.MaxConns = int32(pgCfg.MaxConnections)
	}

	if pgCfg.MinConnections > 0 {
		poolCfg.MinConns = int32(pgCfg.MinConnections)
	}

	if pgCfg.MaxLifetime > 0 {
		poolCfg.MaxConnLifetime = pgCfg.MaxLifetime
	}

	if pgCfg.MaxIdleTime > 0 {
		poolCfg.MaxConnIdleTime = pgCfg.MaxIdleTime
	}

	if pgCfg.HealthCheckPeriod > 0 {
		poolCfg.HealthCheckPeriod = pgCfg.HealthCheckPeriod
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}
