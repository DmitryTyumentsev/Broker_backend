package config

import (
	"errors"
	"fmt"
	"time"

	sharedconfig "Donate_backend/shared/pkg/config"
)

const serviceName = "authservice"

type Config struct {
	Business Business `mapstructure:"business"`
	Database Database `mapstructure:"database"`
	Server   Server   `mapstructure:"server"`
}

type Business struct {
	ContextTimeout time.Duration `mapstructure:"context_timeout"`
}

type Database struct {
	Postgres Postgres `mapstructure:"postgres"`
}

type Postgres struct {
	Dsn               string        `mapstructure:"dsn"`
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port"`
	Username          string        `mapstructure:"username"`
	Password          string        `mapstructure:"password"`
	DatabaseName      string        `mapstructure:"database_name"`
	SSLMode           string        `mapstructure:"ssl_mode"`
	MigrationsPath    string        `mapstructure:"migrations_path"`
	MigrationsTable   string        `mapstructure:"migrations_table"`
	SchemaName        string        `mapstructure:"schema_name"`
	MaxConnections    int           `mapstructure:"max_connections"`
	MinConnections    int           `mapstructure:"min_connections"`
	MaxIdleTime       time.Duration `mapstructure:"max_idle_time"`
	MaxLifetime       time.Duration `mapstructure:"max_lifetime"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	ConnectTimeout    time.Duration `mapstructure:"connect_timeout"`
}

type Server struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func LoadConfig() (*Config, error) {
	cfg := new(Config)

	if err := sharedconfig.Load(serviceName, cfg); err != nil {
		return nil, fmt.Errorf("load authservice config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Host == "" {
		return errors.New("server.host is required")
	}
	if c.Server.Port <= 0 {
		return errors.New("server.port must be > 0")
	}
	if c.Business.ContextTimeout <= 0 {
		return errors.New("business.context_timeout must be > 0")
	}

	pg := c.Database.Postgres
	if pg.Host == "" && pg.Dsn == "" {
		return errors.New("database.postgres.host or database.postgres.dsn is required")
	}
	if pg.Port < 0 {
		return errors.New("database.postgres.port must be >= 0")
	}
	if pg.MaxConnections <= 0 {
		return errors.New("database.postgres.max_connections must be > 0")
	}
	if pg.MinConnections < 0 {
		return errors.New("database.postgres.min_connections must be >= 0")
	}

	return nil
}
