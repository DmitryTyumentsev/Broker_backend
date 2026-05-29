package app

import (
	"Broker_backend/services/authservice/internal/infra/clock"
	"Broker_backend/services/authservice/internal/infra/repositories/postgres"
	"Broker_backend/services/authservice/internal/infra/repositories/postgres/users"
	"Broker_backend/services/authservice/internal/infra/security/jwt"
	"Broker_backend/services/authservice/internal/infra/security/passwordhasher"
	"Broker_backend/services/authservice/internal/transport/grpchandler"
	"Broker_backend/services/authservice/internal/usecases"
	authv1 "Broker_backend/shared/pkg/grpc/gen/auth/v1"

	"google.golang.org/grpc"
)

func InitApp() {
	cfg := loadConfig()
	logger := newLogger()

	pgPool := connectPostgres(cfg)
	redisClient := connectRedis(cfg)

	pg := postgres.NewPostgres(pgPool, cfg)

	usersRepo := users.NewRepository(pg)
	sessionsRepo := refreshsessions.NewRepository(pg)

	passHasher := passwordhasher.NewPasswordHasher()
	accessIssuer := jwt.NewAccessTokenIssuer(cfg, logger)
	refreshTokenService := jwt.NewRefreshTokenService()
	clock := clock.NewRealClock()

	service := usecases.NewService(
		cfg,
		logger,
		usersRepo,
		sessionsRepo,
		passHasher,
		accessIssuer,
		refreshTokenService,
		clock,
	)

	handler := grpchandler.NewHandler(service, logger)

	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, handler)
}
