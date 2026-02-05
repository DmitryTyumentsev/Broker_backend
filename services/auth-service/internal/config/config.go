package config

import (
	"time"
)

const (
	AuthServiceName = "auth"
)

type Config struct {
	Env      string `yaml:"env" env:"ENV"`
	Database `yaml:"database"` //TODO: правильно так или сделать Database *Database? как в памяти все это устроено чтобы было ниже потребление(сколько байт сейчас выделяется и сколько если *Database использовать? Как делать чтобы было удобнее использовать в дальнейшем в коде? Верно понимаю что только при запуске приложения один раз в текущей вариации каждый раз когда вызывается эта структура в любом файле проекта будет создаваться копия в памяти?
	//Cache    `yaml:"cache"`
	Server `yaml:"server"`
} //TODO: должны ли быть go.mod в каждом отдельном микросервисе? зачем они вообще нужны?

type Database struct {
	Postgres `yaml:"postgres"`
}

type Postgres struct {
	// ===== Общие настройки =====
	Dsn             string `yaml:"dsn" env:"DB_DSN"`
	Host            string `yaml:"host" env:"HOST"`
	Port            int    `yaml:"port" env:"PORT"`
	Username        string `yaml:"username" env:"USERNAME"`
	Password        string `yaml:"password" env:"PASSWORD"`
	DatabaseName    string `yaml:"database_name" env:"DATABASE_NAME"`
	SSLMode         string `yaml:"ssl_mode" env:"SSL_MODE" env-default:"disable"`
	MigrationsPath  string `yaml:"migrations_path" env:"MIGRATIONS_PATH"`
	MigrationsTable string `yaml:"migrations_table" env:"MIGRATIONS_TABLE"`
	SchemaName      string `yaml:"schema_name" env:"SCHEMA_NAME"`

	// ==== Настройки всего пула =====
	MaxConnections    *int          `yaml:"max_connections" env:"MAX_CONNECTIONS" env-default:"10"` //TODO: зачем пишем *int? где выигрываем?
	MinConnections    *int          `yaml:"min_connections" env:"MIN_CONNECTIONS" env-default:"1"`
	MaxIdleTime       time.Duration `yaml:"max_idle_time" env:"MAX_IDLE_TIME" env-default:"30m"`
	MaxLifetime       time.Duration `yaml:"max_lifetime" env:"MAX_LIFETIME" env-default:"60m"`
	HealthCheckPeriod time.Duration `yaml:"health_check_period" env:"HEALTH_CHECK_PERIOD" env-default:"5m"`

	// ===== Настройки отдельного соединения
	WriteTimeout   time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" env-default:"30s"`
	ReadTimeout    time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" env-default:"15s"`
	ConnectTimeout time.Duration `yaml:"connect_timeout" env:"CONNECT_TIMEOUT" env-default:"5s"`
}

type Server struct {
	Host              string        `yaml:"host" env:"SERVER_HOST"`
	Port              int           `yaml:"port" env:"PORT"`
	MaxConnections    *int          `yaml:"max_connections" env:"MAX_CONNECTIONS" env-default:"10"`
	MinConnections    *int          `yaml:"min_connections" env:"MIN_CONNECTIONS" env-default:"1"`
	MaxIdleTime       time.Duration `yaml:"max_idle_time" env:"MAX_IDLE_TIME" env-default:"30m"`
	MaxLifetime       time.Duration `yaml:"max_lifetime" env:"MAX_LIFETIME" env-default:"60m"`
	HealthCheckPeriod time.Duration `yaml:"health_check_period" env:"HEALTH_CHECK_PERIOD" env-default:"5m"`
}

func LoadConfig() (*Config, error) {

}
