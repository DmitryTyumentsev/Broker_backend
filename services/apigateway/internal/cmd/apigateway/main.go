package main

import (
	httprouter "Broker_backend/services/apigateway/internal/transport/http"
	"Broker_backend/services/apigateway/internal/transport/http/handlers"
	"Broker_backend/services/apigateway/internal/transport/http/handlers/authhandlers"
	"context"
	"fmt"
	"net"
	"strconv"

	"Broker_backend/services/apigateway/internal/clients/authclient"
	"Broker_backend/services/apigateway/internal/config"

	validate "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
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

	conn, authGRPCClient, err := authclient.NewAuthServiceClient(context.Background(), cfg)
	if err != nil {
		logger.Fatal("connect to auth service", zap.Error(err))
	}
	defer func() {
		_ = conn.Close()
	}()

	authClient := authclient.NewClient(authGRPCClient, cfg)
	validator := validate.New()

	authHandler := authhandlers.NewAuthHandler(logger, authClient, validator)

	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	})

	httprouter.SetupRouter(app, &handlers.Deps{
		Auth: authHandler,
	})

	addr := net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port))

	logger.Info("apigateway started", zap.String("addr", addr))

	if err := app.Listen(addr); err != nil {
		logger.Fatal("apigateway stopped", zap.Error(err))
	}
}
