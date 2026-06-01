package config

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	sharedconfig "Broker_backend/shared/pkg/config"
)

const serviceName = "apigateway"

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Business BusinessConfig `mapstructure:"business"`
	AuthGRPC AuthGRPCConfig `mapstructure:"auth_grpc"`
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
	ContextTimeout        time.Duration   `mapstructure:"context_timeout"`
	AccessTokenSecret     string          `mapstructure:"access_token_secret"`
	AccessTokenIssuer     string          `mapstructure:"access_token_issuer"`
	AuthRateLimit         RateLimitConfig `mapstructure:"auth_rate_limit"`
	DefaultRateLimit      RateLimitConfig `mapstructure:"default_rate_limit"`
	ProtectedAllowedRoles []string        `mapstructure:"protected_allowed_roles"`
	AdminAllowedRoles     []string        `mapstructure:"admin_allowed_roles"`
}

type AuthGRPCConfig struct {
	Address string `mapstructure:"address"`
}

type RateLimitConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	Limit   int64         `mapstructure:"limit"`
	Window  time.Duration `mapstructure:"window"`
	Prefix  string        `mapstructure:"prefix"`
}

type DatabaseConfig struct {
	Redis RedisConfig `mapstructure:"redis"`
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

	if strings.TrimSpace(c.Business.AccessTokenSecret) == "" {
		return errors.New("business.access_token_secret is required")
	}

	if len(strings.TrimSpace(c.Business.AccessTokenSecret)) < 32 {
		return errors.New("business.access_token_secret must be at least 32 bytes")
	}

	if strings.TrimSpace(c.Business.AccessTokenIssuer) == "" {
		return errors.New("business.access_token_issuer is required")
	}

	if err := validateRateLimit("business.auth_rate_limit", c.Business.AuthRateLimit); err != nil {
		return err
	}

	if err := validateRateLimit("business.default_rate_limit", c.Business.DefaultRateLimit); err != nil {
		return err
	}

	if err := c.validateAllowedRoles(); err != nil {
		return err
	}

	if strings.TrimSpace(c.AuthGRPC.Address) == "" {
		return errors.New("auth_grpc.address is required")
	}

	if c.RateLimitEnabled() {
		if err := c.validateRedis(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) RateLimitEnabled() bool {
	if c == nil {
		return false
	}

	return c.Business.AuthRateLimit.Enabled || c.Business.DefaultRateLimit.Enabled
}

func (c *Config) validateAllowedRoles() error {
	protectedRoles := nonEmptyStrings(c.Business.ProtectedAllowedRoles)
	if len(protectedRoles) == 0 {
		return errors.New("business.protected_allowed_roles must contain at least one role")
	}

	adminRoles := nonEmptyStrings(c.Business.AdminAllowedRoles)
	if len(adminRoles) == 0 {
		return errors.New("business.admin_allowed_roles must contain at least one role")
	}

	protected := make(map[string]struct{}, len(protectedRoles))
	for _, role := range protectedRoles {
		protected[role] = struct{}{}
	}

	for _, role := range adminRoles {
		if _, ok := protected[role]; !ok {
			return fmt.Errorf(
				"business.admin_allowed_roles contains role %q that is not in business.protected_allowed_roles",
				role,
			)
		}
	}

	return nil
}

func (c *Config) validateRedis() error {
	redis := c.Database.Redis
	if strings.TrimSpace(redis.Host) == "" {
		return errors.New("database.redis.host is required")
	}

	if redis.Port <= 0 {
		return errors.New("database.redis.port must be positive")
	}

	if redis.DB < 0 {
		return errors.New("database.redis.db must not be negative")
	}

	if redis.DialTimeout <= 0 {
		return errors.New("database.redis.dial_timeout must be positive")
	}

	if redis.ReadTimeout <= 0 {
		return errors.New("database.redis.read_timeout must be positive")
	}

	if redis.WriteTimeout <= 0 {
		return errors.New("database.redis.write_timeout must be positive")
	}

	return nil
}

func nonEmptyStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}

	return out
}

func validateRateLimit(name string, cfg RateLimitConfig) error {
	if !cfg.Enabled {
		return nil
	}

	if cfg.Limit <= 0 {
		return fmt.Errorf("%s.limit must be positive", name)
	}

	if cfg.Window <= 0 {
		return fmt.Errorf("%s.window must be positive", name)
	}

	return nil
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
