package postgres

import (
	"Donate_backend/services/authservice/internal/config"
	"context"
)

type Postgres struct { //TODO: где правильнее хранить структуру Postgres чтобы вызывать ее? как по мне,
	// лучше хранить ее в пакете postgres а в пакетах где сами методы бд, создавать локально структуру и
	//в нее передавать структуру postgres(так написал щас сам, посмотри чуть ниже). Так ли делают на проектах high load?
	//и второй вопрос по неймингу - в пакетах где методы бд, я называю структуру также Postgres как тут. Принято ли так?
	//и еще вопрос - по логгеру - я знаю что логировать нужно в одном месте. То есть если у нас в бд ошибка - залогируй только там,
	//если в сервисе - только там, не нужно одно и тоже место логировать в каждом слое. А как мы поймем была ошибка на предыдущем
	//слое залогирована или нет?
	Pool   *pgxpool.Pool
	Logger *zap.Logger
	Cfg    *config.Config
}

func NewPostgres(pool *pgxpool.Pool, logger *zap.Logger, cfg *config.Config) *Postgres {
	return &Postgres{
		Pool:   pool,
		Logger: logger,
		Cfg:    cfg,
	}
}

func (pg *Postgres) WriteWithTimeout(ctx context.Context, query string) error {
	ctx, cancel := context.WithTimeout(pg.Cfg.Database.Postgres.WriteTimeout)
	defer cancel()
	return pg.Pool.Exec(ctx, query)
}

func (pg *Postgres) ReadRowWithTimeout(ctx context.Context, query string) (interface{}, error) {
	//TODO: принято в entity возвращать данные сразу или создается промежуточная структура?
	ctx, cancel := context.WithTimeout(pg.Cfg.Database.Postgres.ReadTimeout)
	defer cancel()
	return pg.Pool.QueryRow(ctx, query)

	//TODO: расскажи в каких случаях нам достаточно QueryRow, а в каких QueryRows?
}

func (pg *Postgres) ReadRowsWithTimeout(ctx context.Context, query string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(pg.Cfg.Database.Postgres.ReadTimeout)
	defer cancel()
	return pg.Pool.QueryRows(ctx, query) //TODO: тут другой пакет используется вместо pgxpool?
}
