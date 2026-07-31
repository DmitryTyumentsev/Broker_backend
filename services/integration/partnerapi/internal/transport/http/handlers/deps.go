package handlers

import (
	"Broker_backend/services/integration/partnerapi/internal/config"
	"Broker_backend/services/integration/partnerapi/internal/transport/http/handlers/authhandlers"
	"Broker_backend/services/integration/partnerapi/internal/transport/http/handlers/brokerhandlers"
	middleware2 "Broker_backend/services/integration/partnerapi/internal/transport/http/middleware"
	"errors"

	validate "github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Deps struct {
	Auth           *authhandlers.AuthHandler
	Broker         *brokerhandlers.BrokerHandler
	Config         *config.Config
	Logger         *zap.Logger
	Redis          *redis.Client
	Validator      *validate.Validate
	AccessVerifier middleware2.AccessTokenVerifier
	Metrics        *middleware2.PrometheusMetrics
	Authz          middleware2.AuthzPolicy
}

func (d *Deps) Validate() error {
	switch {
	case d == nil:
		return errors.New("http handlers deps are nil")
	case d.Auth == nil:
		return errors.New("auth handler is required")
	case d.Broker == nil:
		return errors.New("brokerservice handler is required")
	case d.Config == nil:
		return errors.New("config is required")
	case d.Logger == nil:
		return errors.New("logger is required")
	case d.Validator == nil:
		return errors.New("validator is required")
	case d.AccessVerifier == nil:
		return errors.New("access verifier is required")
	case d.Metrics == nil:
		return errors.New("metrics registry is required")
	case d.Authz == nil:
		return errors.New("authorization policy is required")
	case d.Redis == nil && d.redisRequired():
		return errors.New("redis client is required when redis-backed middleware is enabled")
	}

	if err := d.Auth.Validate(); err != nil {
		return err
	}

	return d.Broker.Validate()
}

func (d *Deps) redisRequired() bool {
	if d == nil || d.Config == nil {
		return false
	}

	return d.Config.RedisRequired()
}
