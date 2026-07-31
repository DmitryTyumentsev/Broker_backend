package app

import (
	"Broker_backend/services/app/authservice/internal/config"
	jwt2 "Broker_backend/services/app/authservice/internal/infra/security/jwt"
	"Broker_backend/services/app/authservice/internal/infra/security/passwordhasher"
	"Broker_backend/services/app/authservice/internal/repository/postgres"
	"Broker_backend/services/app/authservice/internal/repository/postgres/sessions"
	"Broker_backend/services/app/authservice/internal/repository/postgres/users"
	grpctransport "Broker_backend/services/app/authservice/internal/transport/grpc"
	"Broker_backend/services/app/authservice/internal/usecase"
	"Broker_backend/shared/pkg/clock"
	"context"
	"fmt"
	"net"
	"time"

	authv1 "Broker_backend/gen/auth/v1"
	grpcobservability "Broker_backend/shared/pkg/grpc/observability"
	sharedtracing "Broker_backend/shared/pkg/tracing"

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

	tracingCtx, cancelTracing := context.WithTimeout(rootCtx, cfg.Business.ContextTimeout)
	tracerProvider, err := sharedtracing.InitTracerProvider(tracingCtx, sharedtracing.Config{
		Enabled:      cfg.Observability.Tracing.Enabled,
		ServiceName:  cfg.Observability.Tracing.ServiceName,
		OTLPEndpoint: cfg.Observability.Tracing.OTLPEndpoint,
		Insecure:     cfg.Observability.Tracing.Insecure,
		SampleRatio:  cfg.Observability.Tracing.SampleRatio,
	})
	cancelTracing()
	if err != nil {
		logger.Fatal("init tracing failed", zap.Error(err))
	}
	defer func() {
		shutdownCtx, cancelShutdown := context.WithTimeout(rootCtx, cfg.Business.ContextTimeout)
		defer cancelShutdown()
		_ = tracerProvider.Shutdown(shutdownCtx)
	}()

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

	accessIssuer, err := jwt2.NewAccessTokenIssuer(cfg, logger)
	if err != nil {
		logger.Fatal("create access token issuer failed", zap.Error(err))
	}

	refreshTokenService := jwt2.NewRefreshTokenService()
	realClock := clock.NewRealClock()

	authService := usecase.NewService(
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

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcobservability.TraceUnaryServerInterceptor(cfg.Observability.Tracing.ServiceName),
			unaryContextTimeout(cfg.Business.ContextTimeout),
		),
	)

	authv1.RegisterAuthServiceServer(
		grpcServer,
		grpctransport.NewHandler(authService, logger),
	)

	logger.Info("authservice grpc started", zap.String("addr", addr))

	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("authservice grpc stopped", zap.Error(err))
	}
}

func unaryContextTimeout(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if timeout <= 0 {
			return handler(ctx, req)
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return handler(ctx, req)
	}
}
