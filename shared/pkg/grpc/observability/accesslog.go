package observability

import (
	"context"
	"time"

	"Broker_backend/shared/pkg/requestctx"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// AccessLogUnaryServerInterceptor пишет по строке на каждый RPC.
//
// Зачем он нужен отдельно от логов в хендлерах. Когда приходит жалоба
// «у брокера 500», первый вопрос — дошёл ли запрос до этого сервиса и
// с каким кодом он ушёл обратно. Ответ должен быть в логе всегда, даже
// если конкретный хендлер не залогировал ничего.
//
// request_id приезжает из partnerapi в метаданных и кладётся в контекст
// TraceUnaryServerInterceptor'ом. Именно по нему сшиваются строчки из
// двух сервисов: без него в логе fixationservice невозможно найти ту же
// операцию, что видел клиент.
//
// Ставить ПОСЛЕ TraceUnaryServerInterceptor в цепочке: до него в контексте
// ещё нет ни request_id, ни trace_id.
func AccessLogUnaryServerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		startedAt := time.Now()

		resp, err := handler(ctx, req)

		requestID, _ := requestctx.RequestIDFromContext(ctx)

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.String("code", status.Code(err).String()),
			zap.Duration("duration", time.Since(startedAt)),
			zap.String("request_id", requestID),
			zap.String("trace_id", traceIDFromContext(ctx)),
		}

		// Уровень по коду, а не по факту ошибки: отказ бизнес-правила —
		// это штатная работа сервиса, и Error на него шумит в алертах.
		// Что именно пошло не так, пишет хендлер; здесь только исход.
		if err != nil {
			logger.Warn("grpc access", append(fields, zap.Error(err))...)

			return resp, err
		}

		logger.Info("grpc access", fields...)

		return resp, nil
	}
}

func traceIDFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.HasTraceID() {
		return ""
	}

	return spanCtx.TraceID().String()
}
