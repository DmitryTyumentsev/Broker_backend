package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	sharedconfig "Broker_backend/shared/pkg/config"
)

const serviceName = "authservice"

type Config struct {
	Environment   string              `mapstructure:"environment"`
	Server        ServerConfig        `mapstructure:"server"`
	Business      BusinessConfig      `mapstructure:"business"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	Database      DatabaseConfig      `mapstructure:"database"`
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

	AccessTokenAlg    string `mapstructure:"access_token_alg"`
	AccessTokenType   string `mapstructure:"access_token_type"`
	AccessTokenSecret string `mapstructure:"access_token_secret"`
	AccessTokenIssuer string `mapstructure:"access_token_issuer"`
}

type ObservabilityConfig struct {
	Tracing TracingConfig `mapstructure:"tracing"`
}

type TracingConfig struct {
	Enabled      bool    `mapstructure:"enabled"`
	ServiceName  string  `mapstructure:"service_name"`
	OTLPEndpoint string  `mapstructure:"otlp_endpoint"`
	Insecure     bool    `mapstructure:"insecure"`
	SampleRatio  float64 `mapstructure:"sample_ratio"`
}

type DatabaseConfig struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
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

type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
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
	if c == nil {
		return errors.New("config is nil")
	}

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

	if strings.TrimSpace(c.Business.AccessTokenAlg) == "" {
		return errors.New("business.access_token_alg is required")
	}

	if strings.TrimSpace(c.Business.AccessTokenType) == "" {
		return errors.New("business.access_token_type is required")
	}

	if strings.TrimSpace(c.Business.AccessTokenSecret) == "" {
		return errors.New("business.access_token_secret is required")
	}

	if len(strings.TrimSpace(c.Business.AccessTokenSecret)) < 32 {
		return errors.New("business.access_token_secret must be at least 32 bytes")
	}

	if strings.TrimSpace(c.Business.AccessTokenIssuer) == "" {
		return errors.New("business.access_token_issuer is required")
	}

	if err := validateTracing(c.Observability.Tracing); err != nil {
		return err
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

	if pg.ConnectTimeout <= 0 {
		return errors.New("database.postgres.connect_timeout must be positive")
	}

	if pg.ReadTimeout <= 0 {
		return errors.New("database.postgres.read_timeout must be positive")
	}

	if pg.WriteTimeout <= 0 {
		return errors.New("database.postgres.write_timeout must be positive")
	}

	return nil
}

func validateTracing(cfg TracingConfig) error {
	if !cfg.Enabled {
		return nil
	}

	if strings.TrimSpace(cfg.ServiceName) == "" {
		return errors.New("observability.tracing.service_name is required")
	}

	if strings.TrimSpace(cfg.OTLPEndpoint) == "" {
		return errors.New("observability.tracing.otlp_endpoint is required")
	}

	if cfg.SampleRatio < 0 || cfg.SampleRatio > 1 {
		return errors.New("observability.tracing.sample_ratio must be between 0 and 1")
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

func (r RedisConfig) AddrRedis() string {
	host := strings.TrimSpace(r.Host)
	if host == "" {
		host = "localhost"
	}

	port := r.Port
	if port <= 0 {
		port = 6379
	}

	return net.JoinHostPort(host, strconv.Itoa(port))
}

func (s ServerConfig) AddrServer() string {
	host := strings.TrimSpace(s.Host)
	if host == "" {
		host = "0.0.0.0"
	}

	port := s.Port
	if port <= 0 {
		port = 50051
	}

	return net.JoinHostPort(host, strconv.Itoa(port))
}
