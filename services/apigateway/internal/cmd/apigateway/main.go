package main

import (
	"Broker_backend/services/apigateway/internal/transport/http/handlers/brokerhandlers"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"Broker_backend/services/apigateway/internal/authz"
	"Broker_backend/services/apigateway/internal/clients/authclient"
	"Broker_backend/services/apigateway/internal/clients/brokerclient"
	"Broker_backend/services/apigateway/internal/config"
	redisclient "Broker_backend/services/apigateway/internal/infra/cache/redis"
	httprouter "Broker_backend/services/apigateway/internal/transport/http"
	"Broker_backend/services/apigateway/internal/transport/http/dto"
	"Broker_backend/services/apigateway/internal/transport/http/handlers"
	"Broker_backend/services/apigateway/internal/transport/http/handlers/authhandlers"
	"Broker_backend/services/apigateway/internal/transport/http/middleware"
	sharedjwt "Broker_backend/shared/pkg/security/jwt"
	sharedtracing "Broker_backend/shared/pkg/tracing"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
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

	tracingCtx, cancelTracing := context.WithTimeout(context.Background(), cfg.OperationTimeout())
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
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.OperationTimeout())
		defer cancelShutdown()
		_ = tracerProvider.Shutdown(shutdownCtx)
	}()

	conn, authGRPCClient, err := authclient.NewAuthServiceClient(cfg)
	if err != nil {
		logger.Fatal("create auth grpc client failed", zap.Error(err))
	}
	defer func() {
		_ = conn.Close()
	}()

	var redisClient *redis.Client
	if cfg.RedisRequired() {
		redisCtx, cancelRedis := context.WithTimeout(context.Background(), cfg.OperationTimeout())

		redisClient, err = redisclient.NewClient(redisCtx, cfg.Database.Redis)
		cancelRedis()
		if err != nil {
			logger.Fatal("create redis client failed", zap.Error(err))
		}
		defer func() {
			_ = redisClient.Close()
		}()
	}

	accessVerifier, err := sharedjwt.NewAccessTokenVerifier(sharedjwt.AccessTokenVerifierConfig{
		Secret: cfg.Business.AccessTokenSecret,
		Issuer: cfg.Business.AccessTokenIssuer,
	})
	if err != nil {
		logger.Fatal("create access token verifier failed", zap.Error(err))
	}

	authClient, err := authclient.NewClient(authGRPCClient, cfg)
	if err != nil {
		logger.Fatal("create auth client failed", zap.Error(err))
	}

	authzPolicy, err := authz.NewRolePermissionPolicy(cfg.Business.Authz)
	if err != nil {
		logger.Fatal("create authorization policy failed", zap.Error(err))
	}

	brokerConn, brokerGRPCClient, err := brokerclient.NewBrokerServiceClient(cfg)
	if err != nil {
		logger.Fatal("create broker grpc client failed", zap.Error(err))
	}
	defer func() { _ = brokerConn.Close() }()

	brokerClient, err := brokerclient.NewClient(brokerGRPCClient, cfg)
	if err != nil {
		logger.Fatal("create broker client failed", zap.Error(err))
	}

	validator := middleware.NewRequestValidator()

	authHandler := authhandlers.NewAuthHandler(logger, authClient, validator)
	brokerHandler := brokerhandlers.NewBrokerHandler(logger, brokerClient, validator)

	metrics := middleware.NewPrometheusMetrics("apigateway")

	app := fiber.New(fiber.Config{
		ReadTimeout:           cfg.Server.ReadTimeout,
		WriteTimeout:          cfg.Server.WriteTimeout,
		IdleTimeout:           cfg.Server.IdleTimeout,
		BodyLimit:             cfg.BodyLimitBytes(),
		DisableStartupMessage: false,
		ErrorHandler:          fiberErrorHandler,
	})

	if err := httprouter.SetupRouter(app, &handlers.Deps{
		Auth:           authHandler,
		Broker:         brokerHandler,
		Config:         cfg,
		Logger:         logger,
		Redis:          redisClient,
		Validator:      validator,
		AccessVerifier: accessVerifier,
		Metrics:        metrics,
		Authz:          authzPolicy,
	}); err != nil {
		logger.Fatal("setup router failed", zap.Error(err))
	}

	addr := net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port))

	logger.Info("apigateway started", zap.String("addr", addr))

	if err := app.Listen(addr); err != nil {
		logger.Fatal("apigateway stopped", zap.Error(err))
	}
}

func fiberErrorHandler(c *fiber.Ctx, err error) error {
	statusCode := fiber.StatusInternalServerError
	message := "internal server error"

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		statusCode = fiberErr.Code
		message = fiberErr.Message
	}

	return c.Status(statusCode).JSON(dto.ErrorResponse{
		Code:    statusCode,
		Message: message,
	})
}
