package postgres

import (
	"Broker_backend/services/brokerservice/internal/config"
	"context"

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

func (p *Postgres) DB() *pgxpool.Pool {
	return p.pool
}

func (p *Postgres) WriteWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) { //я трою над этими тремя методами, стоит делать интерфейс или нет, не понимаю. Мне тут лучше в Querier интерфейс добавить эти три метода или в структуру Querier добавить поле cfg а Postgres удалить? как рассуждать, как делать? об этом я и говорю что не понимаю, часто ступор такой в таких ситуациях. Не понимаю мне надо такие кейсы лайвкодом отрабатывать и параллельно фичами или теории не хватает или другой какой-то вариант
	return context.WithTimeout(ctx, p.cfg.Database.Postgres.WriteTimeout)
}

func (p *Postgres) ReadWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, p.cfg.Database.Postgres.ReadTimeout)
}
