package config

import (
	"errors"
	"fmt"
	"time"

	sharedconfig "Broker_backend/shared/pkg/config"
)

const serviceName = "apigateway"

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Business BusinessConfig `mapstructure:"business"`
	AuthGRPC AuthGRPCConfig `mapstructure:"auth_grpc"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type BusinessConfig struct {
	ContextTimeout time.Duration `mapstructure:"context_timeout"`
}

type AuthGRPCConfig struct {
	Address string `mapstructure:"address"`
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

	if c.AuthGRPC.Address == "" {
		return errors.New("auth_grpc.address is required")
	}

	return nil
}
