package config

import (
	"Donate_backend/shared/pkg/config"
	"fmt"
	"log"
	"time"
)

const (
	authServiceName = "auth"
)

type Config struct {
	//TODO: верный ли подход задавать через переменные окружения environment?
	Database `mapstructure:"database" yaml:"database"` //TODO: правильно так или сделать Database *Database? как в памяти все это устроено чтобы было ниже потребление(сколько байт сейчас выделяется и сколько если *Database использовать? Как делать чтобы было удобнее использовать в дальнейшем в коде? Верно понимаю что только при запуске приложения один раз в текущей вариации каждый раз когда вызывается эта структура в любом файле проекта будет создаваться копия в памяти?
	//Cache    `yaml:"cache"`
	Server `mapstructure:"server" yaml:"server"` //TODO: есть ли разница в порядке тегов тут? можно ставить вторым mapstructure или обязательным первым? есть ли приоритетность какая-то в этом?
} //TODO: должны ли быть go.mod в каждом отдельном микросервисе? зачем они вообще нужны?

type Database struct {
	Postgres `mapstructure:"postgres" yaml:"postgres"` //TODO: почему mapstructure копирует yaml? зачем вообще нужен mapstructure? вот есть у нас внутренняя мапа вайпера в которую он собрал данные из yaml, env, переменных окружения. он ее парсит по этим тегам? соотносит что ключ во внутренней мапе и в mapstructure совпал, поэтому записывает значение?
}

type Postgres struct {
	// ===== Общие настройки =====
	Dsn             string `mapstructure:"dsn" yaml:"dsn" env:"DB_DSN"`
	Host            string `mapstructure:"dsn" yaml:"host" env:"HOST"`
	Port            int    `mapstructure:"dsn" yaml:"port" env:"PORT"`
	Username        string `mapstructure:"dsn" yaml:"username" env:"USERNAME"`
	Password        string `mapstructure:"dsn" yaml:"password" env:"PASSWORD"`
	DatabaseName    string `mapstructure:"dsn" yaml:"database_name" env:"DATABASE_NAME"`
	SSLMode         string `mapstructure:"dsn" yaml:"ssl_mode" env:"SSL_MODE" env-default:"disable"`
	MigrationsPath  string `mapstructure:"dsn" yaml:"migrations_path" env:"MIGRATIONS_PATH"`
	MigrationsTable string `mapstructure:"dsn" yaml:"migrations_table" env:"MIGRATIONS_TABLE"`
	SchemaName      string `mapstructure:"dsn" yaml:"schema_name" env:"SCHEMA_NAME"`

	// ==== Настройки всего пула =====
	MaxConnections    *int          `mapstructure:"dsn" yaml:"max_connections" env:"MAX_CONNECTIONS" env-default:"10"` //TODO: зачем пишем *int? где выигрываем?
	MinConnections    *int          `mapstructure:"dsn" yaml:"min_connections" env:"MIN_CONNECTIONS" env-default:"1"`
	MaxIdleTime       time.Duration `mapstructure:"dsn" yaml:"max_idle_time" env:"MAX_IDLE_TIME" env-default:"30m"`
	MaxLifetime       time.Duration `mapstructure:"dsn" yaml:"max_lifetime" env:"MAX_LIFETIME" env-default:"60m"`
	HealthCheckPeriod time.Duration `mapstructure:"dsn" yaml:"health_check_period" env:"HEALTH_CHECK_PERIOD" env-default:"5m"`

	// ===== Настройки отдельного соединения
	WriteTimeout   time.Duration `mapstructure:"dsn" yaml:"write_timeout" env:"WRITE_TIMEOUT" env-default:"30s"`
	ReadTimeout    time.Duration `mapstructure:"dsn" yaml:"read_timeout" env:"READ_TIMEOUT" env-default:"15s"`
	ConnectTimeout time.Duration `mapstructure:"dsn" yaml:"connect_timeout" env:"CONNECT_TIMEOUT" env-default:"5s"`
}

type Server struct {
	Host              string        `mapstructure:"dsn" yaml:"host" env:"SERVER_HOST"`
	Port              int           `mapstructure:"dsn" yaml:"port" env:"PORT"`
	MaxConnections    *int          `mapstructure:"dsn" yaml:"max_connections" env:"MAX_CONNECTIONS" env-default:"10"`
	MinConnections    *int          `mapstructure:"dsn" yaml:"min_connections" env:"MIN_CONNECTIONS" env-default:"1"`
	MaxIdleTime       time.Duration `mapstructure:"dsn" yaml:"max_idle_time" env:"MAX_IDLE_TIME" env-default:"30m"`
	MaxLifetime       time.Duration `mapstructure:"dsn" yaml:"max_lifetime" env:"MAX_LIFETIME" env-default:"60m"`
	HealthCheckPeriod time.Duration `mapstructure:"dsn" yaml:"health_check_period" env:"HEALTH_CHECK_PERIOD" env-default:"5m"`
}

func LoadConfig() (*Config, error) {
	const op = "config.LoadConfig"
	cfg := new(Config)
	loader, err := config.NewViperLoader(cfg, authServiceName)
	log.Printf("cfg: %v, loader: %v", cfg, loader) //TODO: удалить после дебага
	if err != nil {
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	out, ok := loader.ConfigService.(*Config)
	if !ok {
		return nil, fmt.Errorf("op: %s, err: config is not of type *Config", op)
	}
	return out, nil
}
