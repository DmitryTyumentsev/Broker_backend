package infra

import (
	"Broker_backend/services/authservice/internal/config"
	"Broker_backend/services/authservice/internal/infra/repositories/postgres"
	"Broker_backend/services/authservice/internal/infra/security/jwt"
	"context"

	"go.uber.org/zap"
)

type Deps struct {
	logger *zap.Logger
	cfg    *config.Config
}

func (d *Deps) Init(ctx context.Context) error {
	pool, err := postgres.NewConnect(ctx, d.cfg.Database.Postgres)
	if err != nil {
		d.logger.Fatal("connect postgres failed", zap.Error(err))
	}

	postgres := postgres.NewPostgres(pool, d.cfg)

	accessTokenIssuer, err := jwt.NewAccessTokenIssuer(d.cfg, d.logger)
	if err != nil {
		d.logger.Fatal("create access token issuer failed", zap.Error(err))
	}

	refreshTokenService := jwt.NewRefreshTokenService()
}
