package main

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"Donate_backend/services/authservice/internal/config"
	"Donate_backend/services/authservice/internal/infra/cache/redis"
	"Donate_backend/services/authservice/internal/transport/grpchandler"
	authv1 "Donate_backend/shared/pkg/grpc/gen/auth/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
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

	startupCtx, cancel := context.WithTimeout(context.Background(), cfg.Database.Redis.DialTimeout)
	defer cancel()

	redisClient, err := redis.NewClient(startupCtx, cfg.Database.Redis)
	if err != nil {
		logger.Fatal("connect redis failed", zap.Error(err))
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("close redis client failed", zap.Error(err))
		}
	}()

	logger.Info(
		"redis connected",
		zap.String("addr", cfg.Database.Redis.Addr()),
		zap.Int("db", cfg.Database.Redis.DB),
	)

	addr := net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port))

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("listen tcp", zap.String("addr", addr), zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	authv1.RegisterAuthServiceServer(
		grpcServer,
		grpchandler.NewHandler(nil),
	)

	logger.Info("authservice grpc started", zap.String("addr", addr))

	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("authservice grpc stopped", zap.Error(err))
	}
}
