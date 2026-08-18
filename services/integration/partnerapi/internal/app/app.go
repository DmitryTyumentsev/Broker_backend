// ═══════════════════════════════════════════════════════════════════════
//  COMPOSITION ROOT partnerapi — единственное место, где создаются
//  конкретные типы и подставляются в интерфейсы.
//
//  Читай пометки:
//    [БОЙЛЕРПЛЕЙТ] — пишется один раз на сервис, при новой фиче не трогаешь
//    [ФИЧА]        — сюда добавляешь строки под каждую новую ручку
// ═══════════════════════════════════════════════════════════════════════

package app

import (
	"Broker_backend/services/integration/partnerapi/internal/authz"
	"Broker_backend/services/integration/partnerapi/internal/clients/authclient"
	"Broker_backend/services/integration/partnerapi/internal/clients/fixationclient"
	"Broker_backend/services/integration/partnerapi/internal/config"
	redisclient "Broker_backend/services/integration/partnerapi/internal/infra/redis"
	httprouter "Broker_backend/services/integration/partnerapi/internal/transport"
	"Broker_backend/services/integration/partnerapi/internal/transport/dto"
	"Broker_backend/services/integration/partnerapi/internal/transport/handlers"
	"Broker_backend/services/integration/partnerapi/internal/transport/handlers/authhandlers"
	"Broker_backend/services/integration/partnerapi/internal/transport/handlers/fixationhandlers"
	"Broker_backend/services/integration/partnerapi/internal/transport/middleware"
	"context"
	"errors"
	"fmt"
	"net"
	"os/signal"
	"strconv"
	"syscall"

	sharedlogger "Broker_backend/shared/pkg/logger"
	sharedjwt "Broker_backend/shared/pkg/security/jwt"
	sharedtracing "Broker_backend/shared/pkg/tracing"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Run поднимает partnerapi и блокируется до сигнала остановки.
// Возвращает error, а не паникует: решение «что делать с ошибкой»
// принимает main, а не библиотечный код.
func Run() error {
	// ── [БОЙЛЕРПЛЕЙТ] 1. Конфиг ────────────────────────────────────────
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// ── [БОЙЛЕРПЛЕЙТ] 2. Логгер ────────────────────────────────────────
	logger, err := sharedlogger.New(cfg.Environment)
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}
	defer func() { _ = logger.Sync() }()

	// ── [БОЙЛЕРПЛЕЙТ] 3. Корневой контекст + перехват SIGINT/SIGTERM ───
	// Без него `docker stop` доходит до SIGKILL и обрывает запросы в полёте.
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ── [БОЙЛЕРПЛЕЙТ] 4. Трейсинг ──────────────────────────────────────
	tracingCtx, cancelTracing := context.WithTimeout(rootCtx, cfg.OperationTimeout())
	tracerProvider, err := sharedtracing.InitTracerProvider(tracingCtx, sharedtracing.Config{
		Enabled:      cfg.Observability.Tracing.Enabled,
		ServiceName:  cfg.Observability.Tracing.ServiceName,
		OTLPEndpoint: cfg.Observability.Tracing.OTLPEndpoint,
		Insecure:     cfg.Observability.Tracing.Insecure,
		SampleRatio:  cfg.Observability.Tracing.SampleRatio,
	})
	cancelTracing()
	if err != nil {
		return fmt.Errorf("init tracing: %w", err)
	}
	defer func() {
		// context.WithoutCancel: rootCtx на этот момент уже отменён сигналом,
		// а спаны надо успеть доотправить.
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(rootCtx), cfg.OperationTimeout())
		defer cancel()
		_ = tracerProvider.Shutdown(shutdownCtx)
	}()

	// ── [БОЙЛЕРПЛЕЙТ] 5. Клиенты к соседним сервисам ───────────────────
	authConn, authGRPCClient, err := authclient.NewAuthServiceClient(cfg)
	if err != nil {
		return fmt.Errorf("create authservice grpc client: %w", err)
	}
	defer func() { _ = authConn.Close() }()

	authClient, err := authclient.NewClient(authGRPCClient, cfg)
	if err != nil {
		return fmt.Errorf("create authservice client: %w", err)
	}

	fixationConn, fixationGRPCClient, err := fixationclient.NewFixationServiceClient(cfg)
	if err != nil {
		return fmt.Errorf("create fixationservice grpc client: %w", err)
	}
	defer func() { _ = fixationConn.Close() }()

	fixationClient, err := fixationclient.NewClient(fixationGRPCClient, cfg)
	if err != nil {
		return fmt.Errorf("create fixationservice client: %w", err)
	}

	// ── [БОЙЛЕРПЛЕЙТ] 6. Redis ─────────────────────────────────────────
	// Нужен только если включены лимиты или идемпотентность — иначе
	// сервис поднимется и без него.
	var redisClient *redis.Client
	if cfg.RedisRequired() {
		redisCtx, cancelRedis := context.WithTimeout(rootCtx, cfg.OperationTimeout())
		redisClient, err = redisclient.NewClient(redisCtx, cfg.Database.Redis)
		cancelRedis()
		if err != nil {
			return fmt.Errorf("create redis client: %w", err)
		}
		defer func() { _ = redisClient.Close() }()
	}

	// ── [БОЙЛЕРПЛЕЙТ] 7. Проверка токенов и политика доступа ───────────
	accessVerifier, err := sharedjwt.NewAccessTokenVerifier(sharedjwt.AccessTokenVerifierConfig{
		Secret: cfg.Business.AccessTokenSecret,
		Issuer: cfg.Business.AccessTokenIssuer,
	})
	if err != nil {
		return fmt.Errorf("create access token verifier: %w", err)
	}

	authzPolicy, err := authz.NewRolePermissionPolicy(cfg.Business.Authz)
	if err != nil {
		return fmt.Errorf("create authorization policy: %w", err)
	}

	validator := middleware.NewRequestValidator()
	metrics := middleware.NewPrometheusMetrics("partnerapi")

	// ═══════════════════════════════════════════════════════════════════
	//  [ФИЧА] 8. Хендлеры. Одна строка на каждый новый набор ручек.
	//  Порядок аргументов везде один: logger, клиент, валидатор.
	// ═══════════════════════════════════════════════════════════════════
	authHandler := authhandlers.NewAuthHandler(logger, authClient, validator)
	fixationHandler := fixationhandlers.NewFixationHandler(logger, fixationClient, validator)

	// ── [БОЙЛЕРПЛЕЙТ] 9. HTTP-сервер ───────────────────────────────────
	app := fiber.New(fiber.Config{
		ReadTimeout:           cfg.Server.ReadTimeout,
		WriteTimeout:          cfg.Server.WriteTimeout,
		IdleTimeout:           cfg.Server.IdleTimeout,
		BodyLimit:             cfg.BodyLimitBytes(),
		DisableStartupMessage: true, // стартовую простыню fiber заменяем своей строкой в лог
		ErrorHandler:          fiberErrorHandler,
	})

	if err := httprouter.SetupRouter(app, &handlers.Deps{
		Auth:           authHandler,
		Fixation:       fixationHandler,
		Config:         cfg,
		Logger:         logger,
		Redis:          redisClient,
		Validator:      validator,
		AccessVerifier: accessVerifier,
		Metrics:        metrics,
		Authz:          authzPolicy,
	}); err != nil {
		return fmt.Errorf("setup router: %w", err)
	}

	// ── [БОЙЛЕРПЛЕЙТ] 10. Запуск + graceful shutdown ───────────────────
	addr := net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port))

	serveErr := make(chan error, 1) // буфер 1: горутина не повиснет, если никто не читает

	go func() {
		logger.Info("partnerapi started", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		// Сервер упал сам (порт занят, listener закрылся).
		if err != nil {
			return fmt.Errorf("http listen: %w", err)
		}
		return nil

	case <-rootCtx.Done():
		// Пришёл SIGTERM. Даём доработать текущим запросам, но не вечно:
		// ShutdownWithContext рвёт соединения по истечении потолка.
		logger.Info("shutdown signal received, stopping gracefully")

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(rootCtx), cfg.RequestTimeout())
		defer cancel()

		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			return fmt.Errorf("http shutdown: %w", err)
		}

		logger.Info("partnerapi stopped gracefully")
		return nil
	}
}

// ── [БОЙЛЕРПЛЕЙТ] вспомогательное ─────────────────────────────────────

// fiberErrorHandler — последний рубеж: всё, что вернулось из цепочки
// хендлеров как error и никем не записано, превращается в наш JSON.
// Без него fiber отдаст текстом и без контракта ошибки.
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
