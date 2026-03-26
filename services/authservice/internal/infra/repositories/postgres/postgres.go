package postgres

import (
	"Donate_backend/services/authservice/internal/config"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct { //TODO: где правильнее хранить структуру Postgres чтобы вызывать ее? как по мне,
	// лучше хранить ее в пакете postgres а в пакетах где сами методы бд, создавать локально структуру и
	//в нее передавать структуру postgres(так написал щас сам, посмотри чуть ниже). Так ли делают на проектах high load?
	//и второй вопрос по неймингу - в пакетах где методы бд, я называю структуру также Postgres как тут. Принято ли так?
	//и еще вопрос - по логгеру - я знаю что логировать нужно в одном месте. То есть если у нас в бд ошибка - залогируй только там,
	//если в сервисе - только там, не нужно одно и тоже место логировать в каждом слое. А как мы поймем была ошибка на предыдущем
	//слое залогирована или нет?
	Pool *pgxpool.Pool
	Cfg  *config.Config
}

func NewPostgres(pool *pgxpool.Pool, cfg *config.Config) *Postgres {
	return &Postgres{
		Pool: pool,
		Cfg:  cfg,
	}
}

func (pg *Postgres) WriteWithTimeout(ctx context.Context, query string) error {
	ctx, cancel := context.WithTimeout(ctx, pg.Cfg.Database.Postgres.WriteTimeout)
	defer cancel()
	_, err := pg.Pool.Exec(ctx, query) //TODO: на что нужно использовать exec? разве не на insert?
	return err
}

func (pg *Postgres) ReadRowWithTimeout(ctx context.Context, query string) (interface{}, error) {
	//TODO: как лучше написать? как написать что возвращает данные? разные же поля могут быть, как мне связать все это?
	ctx, cancel := context.WithTimeout(pg.Cfg.Database.Postgres.ReadTimeout)
	defer cancel()
	err := pg.Pool.QueryRow(ctx, query).Scan()
	return

}

func (pg *Postgres) ReadRowsWithTimeout(ctx context.Context, query string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(pg.Cfg.Database.Postgres.ReadTimeout)
	defer cancel()
	return pg.Pool.QueryRows(ctx, query) //TODO: тут другой пакет используется вместо pgxpool? тот же вопрос как связать параметры на вход и то что возвращает?
}
