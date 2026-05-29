package app

import (
	"context"
	"fmt"
	"net"

	"Broker_backend/services/authservice/internal/config"
	"Broker_backend/services/authservice/internal/infra/clock"
	"Broker_backend/services/authservice/internal/infra/repositories/postgres"
	"Broker_backend/services/authservice/internal/infra/repositories/postgres/sessions"
	"Broker_backend/services/authservice/internal/infra/repositories/postgres/users"
	"Broker_backend/services/authservice/internal/infra/security/jwt"
	"Broker_backend/services/authservice/internal/infra/security/passwordhasher"
	"Broker_backend/services/authservice/internal/transport/grpchandler"
	"Broker_backend/services/authservice/internal/usecases"
	authv1 "Broker_backend/shared/pkg/grpc/gen/auth/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func InitAuthservice() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Errorf("create logger: %w", err))
	}
	defer func() {
		_ = logger.Sync()
	}()

	rootCtx := context.Background()

	postgresCtx, cancelPostgres := context.WithTimeout(rootCtx, cfg.Database.Postgres.ConnectTimeout)
	defer cancelPostgres()

	pool, err := postgres.NewPool(postgresCtx, cfg)
	if err != nil {
		logger.Fatal("connect postgres failed", zap.Error(err))
	}
	defer pool.Close()

	pg := postgres.NewPostgres(pool, cfg)

	usersRepo := users.NewRepository(pg)
	sessionsRepo := sessions.NewRepository(pg)

	passHasher := passwordhasher.NewPasswordHasher()

	accessIssuer, err := jwt.NewAccessTokenIssuer(cfg, logger)
	if err != nil {
		logger.Fatal("create access token issuer failed", zap.Error(err))
	}

	refreshTokenService := jwt.NewRefreshTokenService()
	realClock := clock.NewRealClock()

	authService := usecases.NewService(
		cfg,
		logger,
		usersRepo,
		sessionsRepo,
		passHasher,
		accessIssuer,
		refreshTokenService,
		realClock,
	)

	addr := cfg.Server.AddrServer()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("listen tcp failed", zap.String("addr", addr), zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	authv1.RegisterAuthServiceServer(
		grpcServer,
		grpchandler.NewHandler(authService, logger),
	)

	logger.Info("authservice grpc started", zap.String("addr", addr))

	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("authservice grpc stopped", zap.Error(err))
	}
}
