//  COMPOSITION ROOT — единственное место, где создаются конкретные типы
//  и подставляются в интерфейсы. Всё остальное приложение работает
//  только с интерфейсами и не знает, что под ними.
//
//  Читай пометки:
//    [БОЙЛЕРПЛЕЙТ] — пишется один раз на сервис, при новой фиче не трогаешь
//    [ФИЧА]        — сюда добавляешь строки под каждую новую ручку
// ═══════════════════════════════════════════════════════════════════════

package app

import (
	"Broker_backend/services/integration/fixationservice/internal/config"
	"Broker_backend/services/integration/fixationservice/internal/repository/postgres"
	grpctransport "Broker_backend/services/integration/fixationservice/internal/transport/grpc"
	"Broker_backend/services/integration/fixationservice/internal/usecase"
	"Broker_backend/shared/pkg/clock"
	"context"
	"errors"
	"fmt"
	"net"
	"os/signal"
	"syscall"
	"time"

	fixationv1 "Broker_backend/gen/fixation/v1"
	grpcobservability "Broker_backend/shared/pkg/grpc/observability"
	sharedtracing "Broker_backend/shared/pkg/tracing"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Run поднимает сервис и блокируется до сигнала остановки.
// Возвращает error, а не паникует: решение «что делать с ошибкой»
// принимает main, а не библиотечный код.
func Run() error {
	// ── [БОЙЛЕРПЛЕЙТ] 1. Конфиг ────────────────────────────────────────
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// ── [БОЙЛЕРПЛЕЙТ] 2. Логгер ────────────────────────────────────────
	logger, err := newLogger(cfg)
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}
	defer func() { _ = logger.Sync() }()

	// ── [БОЙЛЕРПЛЕЙТ] 3. Корневой контекст + перехват SIGINT/SIGTERM ───
	// signal.NotifyContext отменяет ctx при Ctrl+C или `docker stop`.
	// Это точка входа graceful shutdown — без него контейнер убивают
	// через SIGKILL и запросы в полёте обрываются.
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ── [БОЙЛЕРПЛЕЙТ] 4. Трейсинг ──────────────────────────────────────
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
		return fmt.Errorf("init tracing: %w", err)
	}
	defer func() {
		// context.WithoutCancel: rootCtx на этот момент уже отменён сигналом,
		// а спаны надо успеть доотправить.
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(rootCtx), cfg.Business.ContextTimeout)
		defer cancel()
		_ = tracerProvider.Shutdown(shutdownCtx)
	}()

	// ── [БОЙЛЕРПЛЕЙТ] 5. Postgres: пул + обёртка + менеджер транзакций ─
	pgCtx, cancelPg := context.WithTimeout(rootCtx, cfg.Database.Postgres.ConnectTimeout)
	defer cancelPg()

	pool, err := postgres.NewPool(pgCtx, cfg)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	pg := postgres.NewPostgres(pool, cfg)
	txManager := postgres.NewTxManager(pool) // ← нужен конструктор, см. txmanager_patch.go

	_ = pg // пригодится, когда появятся репозитории на старой обёртке

	// ── [БОЙЛЕРПЛЕЙТ] 6. Инфраструктурные сервисы ──────────────────────
	realClock := clock.NewRealClock()

	// ═══════════════════════════════════════════════════════════════════
	//  [ФИЧА] 7. Репозитории
	//  Одна строка на каждый новый репозиторий.
	//  Конкретный *Repository здесь подставится в параметр-интерфейс
	//  usecase.FixationRepository ниже — Go проверит методы сам.
	// ═══════════════════════════════════════════════════════════════════
	fixationRepo := postgres.NewRepository(txManager)

	// ═══════════════════════════════════════════════════════════════════
	//  [ФИЧА] 8. Юзкейсы
	//  Сюда добавляешь новые зависимости фичи (репозитории, клиенты
	//  других сервисов, хешеры). Порядок аргументов — как в NewService.
	// ═══════════════════════════════════════════════════════════════════
	fixationService := usecase.NewService(
		cfg,
		logger,
		realClock,    // → интерфейс usecase.Clock
		fixationRepo, // → интерфейс usecase.FixationRepository
		txManager,    // → интерфейс usecase.TxManager
	)

	// ── [БОЙЛЕРПЛЕЙТ] 9. gRPC-сервер ───────────────────────────────────
	addr := cfg.Server.AddrServer()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen tcp %s: %w", addr, err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcobservability.TraceUnaryServerInterceptor(cfg.Observability.Tracing.ServiceName),
			recoveryInterceptor(logger), // паника в хендлере не должна ронять сервис
			unaryContextTimeout(cfg.Business.ContextTimeout),
		),
	)

	// Health-check: k8s ходит сюда для liveness/readiness проб.
	healthSrv := health.NewServer()
	healthv1.RegisterHealthServer(grpcServer, healthSrv)

	// Reflection: позволяет grpcurl/Postman вызывать методы без .proto-файла.
	// Удобно локально; в проде обычно выключают.
	if cfg.Environment == "local" {
		reflection.Register(grpcServer)
	}

	// ═══════════════════════════════════════════════════════════════════
	//  [ФИЧА] 10. Регистрация хендлеров
	//  Одна строка на каждый новый gRPC-сервис из proto.
	// ═══════════════════════════════════════════════════════════════════
	fixationv1.RegisterFixationServiceServer(
		grpcServer,
		grpctransport.NewHandler(fixationv1.UnimplementedFixationServiceServer{}, fixationService, logger),
	)

	// ── [БОЙЛЕРПЛЕЙТ] 11. Запуск + graceful shutdown ───────────────────
	serveErr := make(chan error, 1) // буфер 1: горутина не повиснет, если никто не читает

	go func() {
		logger.Info("fixationservice grpc started", zap.String("addr", addr))
		if err := grpcServer.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		// Сервер упал сам (порт занят, listener закрылся).
		if err != nil {
			return fmt.Errorf("grpc serve: %w", err)
		}
		return nil

	case <-rootCtx.Done():
		// Пришёл SIGTERM. Даём доработать текущим запросам.
		logger.Info("shutdown signal received, stopping gracefully")
		healthSrv.Shutdown() // health начнёт отвечать NOT_SERVING → LB перестанет слать трафик

		stopped := make(chan struct{})
		go func() {
			grpcServer.GracefulStop() // ждёт завершения активных RPC
			close(stopped)
		}()

		select {
		case <-stopped:
			logger.Info("fixationservice stopped gracefully")
		case <-time.After(cfg.Business.ContextTimeout):
			logger.Warn("graceful stop timed out, forcing")
			grpcServer.Stop() // рвём соединения принудительно
		}
		return nil
	}
}

// ── [БОЙЛЕРПЛЕЙТ] вспомогательное ─────────────────────────────────────

func newLogger(cfg *config.Config) (*zap.Logger, error) {
	if cfg.Environment == "local" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction() // JSON-формат: его парсит Loki/ELK
}

// unaryContextTimeout ставит потолок на время обработки одного RPC.
// Клиентский дедлайн приезжает сам (заголовок grpc-timeout); этот —
// страховка на случай, если клиент дедлайн не выставил.
func unaryContextTimeout(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if timeout <= 0 {
			return handler(ctx, req)
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return handler(ctx, req)
	}
}

// recoveryInterceptor ловит панику в хендлере и превращает её в Internal.
// Без него паника в одном RPC роняет весь процесс.
func recoveryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic in grpc handler",
					zap.Any("panic", r),
					zap.String("method", info.FullMethod),
					zap.Stack("stack"),
				)
				err = fmt.Errorf("internal server error")
			}
		}()
		return handler(ctx, req)
	}
}
