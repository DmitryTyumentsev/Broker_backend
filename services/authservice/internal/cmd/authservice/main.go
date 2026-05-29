package main

import (
	"Broker_backend/services/authservice/internal/config"
	"Broker_backend/services/authservice/internal/infra/cache/redis"
	"Broker_backend/services/authservice/internal/transport/grpchandler"
	"Broker_backend/services/authservice/internal/usecases"
	authv1 "Broker_backend/shared/pkg/grpc/gen/auth/v1"
	"context"
	"fmt"
	"net"

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

	businessCtx, cancel := context.WithTimeout(context.Background(), cfg.Business.ContextTimeout)
	defer cancel()

	redisClient, err := redis.NewClient(businessCtx, cfg.Database.Redis) //путаница с контекстами - что они такое, зачем разные, что с чем контексты связывают, какие куда ставить?
	if err != nil {
		logger.Fatal("connect redis failed", zap.Error(err))
	}
	defer func() {
		if err := redisClient.Close(); err != nil { //redis close должен быть тут? зачем вообще нужен Close у редис?
			logger.Error("close redis client failed", zap.Error(err))
		}
	}() //напомни по defer - когда вообще их применять, завязаны ли они на return, в какой момент выполняется то что в defer? зачем нужны, как работают?

	logger.Info(
		"redis connected",
		zap.String("addr", cfg.Database.Redis.AddrRedis()),
		zap.Int("db", cfg.Database.Redis.DB),
	)

	logger.Info("start postgres init")

	//пришли мне метод для коннекта к постгре полный чтобы просто скопировать с учетом моего кода и зависимостей

	//структура самого файла main - как принято - пишу я методы на разных слоях, делаю разные New, которые буду инициировать в main. Мне писать их в блок(слой) к которому относится он(как тут, accessTokenIssuer в слое infra, значит инициирую там где базы) или по другому как-то пишут?

	//как вообще нормально main писать? пишут вообще все сюда или выносят в deps.go по слоям и их уже тут инициируют? или вообще по другому как-то? upd: вынес по итогу из main, давай разберем как правильно
	//а сами методы тоже надо все выписывать? у меня не хватает вызовов методов сейчас для сборки и передачи в них зависимостей

	addrServer := cfg.Server.AddrServer()

	listener, err := net.Listen("tcp", addrServer)
	if err != nil {
		logger.Fatal("listen tcp", zap.String("addr", addrServer), zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	service := usecases.NewService(cfg,
		logger,
	)

	authv1.RegisterAuthServiceServer(
		grpcServer,
		grpchandler.NewHandler(service),
	)

	logger.Info("authservice grpc started", zap.String("addr", addrServer))

	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("authservice grpc stopped", zap.Error(err))
	}

}
