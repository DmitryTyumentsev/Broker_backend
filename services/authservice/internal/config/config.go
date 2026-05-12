package config

import (
	"errors"
	"fmt"
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
	DSN            string        `mapstructure:"dsn"`
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	User           string        `mapstructure:"user"`
	Password       string        `mapstructure:"password"`
	Name           string        `mapstructure:"name"`
	SSLMode        string        `mapstructure:"ssl_mode"`
	MaxConnections int           `mapstructure:"max_connections"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
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

	return nil
}
