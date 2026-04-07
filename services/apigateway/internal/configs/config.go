package configs

import (
	"Donate_backend/shared/pkg/config"
	"fmt"
	"log"
	"time"
)

const (
	apigatewayServiceName = "api-gateway"
)

type Config struct {
	Server   `mapstructure:"server" yaml:"server"`
	Business `mapstructure:"business" yaml:"business"`
}

type Business struct {
	ContextTimeout time.Duration `mapstructure:"context_timeout" yaml:"context_timeout"`
}

type Server struct {
	Host              string        `mapstructure:"host" yaml:"host" env:"SERVER_HOST"`
	Port              int           `mapstructure:"port" yaml:"port" env:"PORT"`
	MaxConnections    *int          `mapstructure:"max_connections" yaml:"max_connections" env:"MAX_CONNECTIONS" env-default:"10"`
	MinConnections    *int          `mapstructure:"min_connections" yaml:"min_connections" env:"MIN_CONNECTIONS" env-default:"1"`
	MaxIdleTime       time.Duration `mapstructure:"max_idle_time" yaml:"max_idle_time" env:"MAX_IDLE_TIME" env-default:"30m"`
	MaxLifetime       time.Duration `mapstructure:"max_lifetime" yaml:"max_lifetime" env:"MAX_LIFETIME" env-default:"60m"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period" yaml:"health_check_period" env:"HEALTH_CHECK_PERIOD" env-default:"5m"`
}

func LoadConfig() (*Config, error) {
	const op = "config.LoadConfig"
	cfg := new(Config)
	loader, err := config.NewViperLoader(cfg, apigatewayServiceName)
	log.Printf("cfg: %v, loader: %v", *cfg, *loader) //TODO: горит желтым *loader, что не так? я же хочу взять данные экземпляра структуры а не данные адреса(указателя). Почему cfg тогда не горит желтым?
	if err != nil {
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	out, ok := loader.ConfigService.(*Config) //TODO: а зачем эта проверка? как оно может не совпать? или это просто подстраховка?
	if !ok {
		return nil, fmt.Errorf("op: %s, err: config is not of type *Config", op)
	}
	return out, nil
}
