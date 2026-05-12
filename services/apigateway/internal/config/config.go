package config

import (
	"errors"
	"fmt"
	"time"

	sharedconfig "Donate_backend/shared/pkg/config"
)

const serviceName = "apigateway"

type Config struct {
	Server   Server   `mapstructure:"server"`
	Business Business `mapstructure:"business"`
	AuthGRPC AuthGRPC `mapstructure:"auth_grpc"`
}

type Server struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type Business struct {
	ContextTimeout time.Duration `mapstructure:"context_timeout"`
}

type AuthGRPC struct {
	Address string `mapstructure:"address"`
}

func LoadConfig() (*Config, error) {
	cfg := new(Config)

	if err := sharedconfig.Load(serviceName, cfg); err != nil {
		return nil, fmt.Errorf("load apigateway config: %w", err)
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
	if c.AuthGRPC.Address == "" {
		return errors.New("auth_grpc.address is required")
	}

	return nil
}
