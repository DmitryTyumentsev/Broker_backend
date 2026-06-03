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
	Server        ServerConfig        `mapstructure:"server"`
	Business      BusinessConfig      `mapstructure:"business"`
	HTTP          HTTPConfig          `mapstructure:"http"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	AuthGRPC      AuthGRPCConfig      `mapstructure:"auth_grpc"`
	Database      DatabaseConfig      `mapstructure:"database"`
}

type ServerConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	BodyLimitBytes int           `mapstructure:"body_limit_bytes"`
}

type BusinessConfig struct {
	ContextTimeout        time.Duration     `mapstructure:"context_timeout"`
	RequestTimeout        time.Duration     `mapstructure:"request_timeout"`
	OperationTimeout      time.Duration     `mapstructure:"operation_timeout"`
	AccessTokenSecret     string            `mapstructure:"access_token_secret"`
	AccessTokenIssuer     string            `mapstructure:"access_token_issuer"`
	AuthRateLimit         RateLimitConfig   `mapstructure:"auth_rate_limit"`
	DefaultRateLimit      RateLimitConfig   `mapstructure:"default_rate_limit"`
	Idempotency           IdempotencyConfig `mapstructure:"idempotency"`
	ProtectedAllowedRoles []string          `mapstructure:"protected_allowed_roles"`
	AdminAllowedRoles     []string          `mapstructure:"admin_allowed_roles"`
}

type HTTPConfig struct {
	CORS            CORSConfig            `mapstructure:"cors"`
	SecurityHeaders SecurityHeadersConfig `mapstructure:"security_headers"`
}

type CORSConfig struct {
	Enabled          bool     `mapstructure:"enabled"`
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAgeSeconds    int      `mapstructure:"max_age_seconds"`
}

type SecurityHeadersConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

type ObservabilityConfig struct {
	Metrics MetricsConfig `mapstructure:"metrics"`
	Tracing TracingConfig `mapstructure:"tracing"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

type TracingConfig struct {
	Enabled      bool    `mapstructure:"enabled"`
	ServiceName  string  `mapstructure:"service_name"`
	OTLPEndpoint string  `mapstructure:"otlp_endpoint"`
	Insecure     bool    `mapstructure:"insecure"`
	SampleRatio  float64 `mapstructure:"sample_ratio"`
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

type IdempotencyConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	TTL              time.Duration `mapstructure:"ttl"`
	Header           string        `mapstructure:"header"`
	LockPrefix       string        `mapstructure:"lock_prefix"`
	MaxResponseBytes int           `mapstructure:"max_response_bytes"`
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

	if c.RequestTimeout() <= 0 {
		return errors.New("business.request_timeout must be positive")
	}

	if c.OperationTimeout() <= 0 {
		return errors.New("business.operation_timeout must be positive")
	}

	if c.BodyLimitBytes() <= 0 {
		return errors.New("server.body_limit_bytes must be positive")
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

	if err := validateIdempotency(c.Business.Idempotency); err != nil {
		return err
	}

	if err := validateCORS(c.HTTP.CORS); err != nil {
		return err
	}

	if err := validateMetrics(c.Observability.Metrics); err != nil {
		return err
	}

	if err := validateTracing(c.Observability.Tracing); err != nil {
		return err
	}

	if err := c.validateAllowedRoles(); err != nil {
		return err
	}

	if strings.TrimSpace(c.AuthGRPC.Address) == "" {
		return errors.New("auth_grpc.address is required")
	}

	if c.RedisRequired() {
		if err := c.validateRedis(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) RequestTimeout() time.Duration {
	if c == nil {
		return 0
	}

	if c.Business.RequestTimeout > 0 {
		return c.Business.RequestTimeout
	}

	return c.Business.ContextTimeout
}

func (c *Config) OperationTimeout() time.Duration {
	if c == nil {
		return 0
	}

	if c.Business.OperationTimeout > 0 {
		return c.Business.OperationTimeout
	}

	return c.Business.ContextTimeout
}

func (c *Config) BodyLimitBytes() int {
	if c == nil {
		return 0
	}

	if c.Server.BodyLimitBytes > 0 {
		return c.Server.BodyLimitBytes
	}

	return 2 * 1024 * 1024
}

func (c *Config) RateLimitEnabled() bool {
	if c == nil {
		return false
	}

	return c.Business.AuthRateLimit.Enabled || c.Business.DefaultRateLimit.Enabled
}

func (c *Config) IdempotencyEnabled() bool {
	if c == nil {
		return false
	}

	return c.Business.Idempotency.Enabled
}

func (c *Config) RedisRequired() bool {
	return c.RateLimitEnabled() || c.IdempotencyEnabled()
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

func validateIdempotency(cfg IdempotencyConfig) error {
	if !cfg.Enabled {
		return nil
	}

	if cfg.TTL <= 0 {
		return errors.New("business.idempotency.ttl must be positive")
	}

	if strings.TrimSpace(cfg.Header) == "" {
		return errors.New("business.idempotency.header is required")
	}

	if strings.TrimSpace(cfg.LockPrefix) == "" {
		return errors.New("business.idempotency.lock_prefix is required")
	}

	if cfg.MaxResponseBytes <= 0 {
		return errors.New("business.idempotency.max_response_bytes must be positive")
	}

	return nil
}

func validateCORS(cfg CORSConfig) error {
	if !cfg.Enabled {
		return nil
	}

	if len(nonEmptyStrings(cfg.AllowOrigins)) == 0 {
		return errors.New("http.cors.allow_origins must contain at least one origin")
	}

	if len(nonEmptyStrings(cfg.AllowMethods)) == 0 {
		return errors.New("http.cors.allow_methods must contain at least one method")
	}

	if len(nonEmptyStrings(cfg.AllowHeaders)) == 0 {
		return errors.New("http.cors.allow_headers must contain at least one header")
	}

	if cfg.MaxAgeSeconds < 0 {
		return errors.New("http.cors.max_age_seconds must not be negative")
	}

	return nil
}

func validateMetrics(cfg MetricsConfig) error {
	if !cfg.Enabled {
		return nil
	}

	if strings.TrimSpace(cfg.Path) == "" {
		return errors.New("observability.metrics.path is required")
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
