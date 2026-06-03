package handlers

import (
	"errors"

	"Broker_backend/services/apigateway/internal/config"
	"Broker_backend/services/apigateway/internal/transport/http/handlers/authhandlers"
	"Broker_backend/services/apigateway/internal/transport/http/middleware"

	validate "github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Deps struct {
	Auth           *authhandlers.AuthHandler
	Config         *config.Config
	Logger         *zap.Logger
	Redis          *redis.Client
	Validator      *validate.Validate
	AccessVerifier middleware.AccessTokenVerifier
	Metrics        *middleware.PrometheusMetrics
}

func (d *Deps) Validate() error {
	switch {
	case d == nil:
		return errors.New("http handlers deps are nil")
	case d.Auth == nil:
		return errors.New("auth handler is required")
	case d.Config == nil:
		return errors.New("config is required")
	case d.Validator == nil:
		return errors.New("validator is required")
	case d.AccessVerifier == nil:
		return errors.New("access verifier is required")
	case d.Metrics == nil:
		return errors.New("metrics registry is required")
	case d.Redis == nil && d.rateLimitEnabled():
		return errors.New("redis client is required when rate limit is enabled")
	default:
		return d.Auth.Validate()
	}
}

func (d *Deps) rateLimitEnabled() bool {
	if d == nil || d.Config == nil {
		return false
	}

	return d.Config.RateLimitEnabled()
}
