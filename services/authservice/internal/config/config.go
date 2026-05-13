package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	sharedconfig "Donate_backend/shared/pkg/config"
)

const serviceName = "authservice"

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Business BusinessConfig `mapstructure:"business"`
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type BusinessConfig struct {
	ContextTimeout       time.Duration `mapstructure:"context_timeout"`
	LifetimeAccessToken  time.Duration `mapstructure:"lifetime_access_token"`
	LifetimeRefreshToken time.Duration `mapstructure:"lifetime_refresh_token"`
}

type DatabaseConfig struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
}

type PostgresConfig struct {
	DSN string `mapstructure:"dsn"`

	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	DatabaseName string `mapstructure:"database_name"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	SSLMode      string `mapstructure:"ssl_mode"`

	MigrationsPath  string `mapstructure:"migrations_path"`
	MigrationsTable string `mapstructure:"migrations_table"`
	SchemaName      string `mapstructure:"schema_name"`

	MaxConnections    int           `mapstructure:"max_connections"`
	MinConnections    int           `mapstructure:"min_connections"`
	MaxIdleTime       time.Duration `mapstructure:"max_idle_time"`
	MaxLifetime       time.Duration `mapstructure:"max_lifetime"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period"`

	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}

	if err := sharedconfig.Load(serviceName, cfg); err != nil {
		return nil, fmt.Errorf("load %s config: %w", serviceName, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate %s config: %w", serviceName, err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return errors.New("server.port must be positive")
	}

	if c.Business.ContextTimeout <= 0 {
		return errors.New("business.context_timeout must be positive")
	}

	if c.Business.LifetimeAccessToken <= 0 {
		return errors.New("business.lifetime_access_token must be positive")
	}

	if c.Business.LifetimeRefreshToken <= 0 {
		return errors.New("business.lifetime_refresh_token must be positive")
	}

	pg := c.Database.Postgres

	if strings.TrimSpace(pg.DSN) == "" {
		if strings.TrimSpace(pg.Host) == "" {
			return errors.New("database.postgres.host is required when dsn is empty")
		}

		if pg.Port <= 0 {
			return errors.New("database.postgres.port must be positive when dsn is empty")
		}

		if strings.TrimSpace(pg.DatabaseName) == "" {
			return errors.New("database.postgres.database_name is required when dsn is empty")
		}

		if strings.TrimSpace(pg.Username) == "" {
			return errors.New("database.postgres.username is required when dsn is empty")
		}
	}

	return nil
}

func (p PostgresConfig) ConnectionString() string {
	if strings.TrimSpace(p.DSN) != "" {
		return p.DSN
	}

	host := strings.TrimSpace(p.Host)
	if host == "" {
		host = "localhost"
	}

	port := p.Port
	if port <= 0 {
		port = 5432
	}

	sslMode := strings.TrimSpace(p.SSLMode)
	if sslMode == "" {
		sslMode = "disable"
	}

	databaseName := strings.TrimPrefix(strings.TrimSpace(p.DatabaseName), "/")

	dsn := url.URL{
		Scheme: "postgres",
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
		Path:   "/" + databaseName,
	}

	if p.Password != "" {
		dsn.User = url.UserPassword(p.Username, p.Password)
	} else {
		dsn.User = url.User(p.Username)
	}

	query := dsn.Query()
	query.Set("sslmode", sslMode)
	dsn.RawQuery = query.Encode()

	return dsn.String()
}

func (p PostgresConfig) GooseTableName() string {
	table := strings.TrimSpace(p.MigrationsTable)
	if table == "" {
		return ""
	}

	schema := strings.TrimSpace(p.SchemaName)
	if schema == "" || schema == "public" || strings.Contains(table, ".") {
		return table
	}

	return schema + "." + table
}
