package middleware

import (
	"time"

	sharedauth "Broker_backend/shared/pkg/auth"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func AccessLog(logger *zap.Logger) fiber.Handler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(c *fiber.Ctx) error {
		startedAt := time.Now()
		err := c.Next()
		duration := time.Since(startedAt)

		statusCode := c.Response().StatusCode()
		if err != nil && statusCode < fiber.StatusBadRequest {
			statusCode = fiber.StatusInternalServerError
		}

		fields := []zap.Field{
			zap.String("request_id", CurrentRequestID(c)),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", statusCode),
			zap.Duration("duration", duration),
			zap.String("ip", c.IP()),
			zap.String("user_agent", c.Get(fiber.HeaderUserAgent)),
			zap.String("trace_id", traceID(c)),
		}

		if principal, ok := sharedauth.PrincipalFromContext(c.UserContext()); ok {
			fields = append(
				fields,
				zap.String("user_id", principal.UserID),
				zap.String("device_id", principal.DeviceID),
				zap.String("role", principal.Role),
			)
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
		}

		logger.Info("http access", fields...)
		return err
	}
}

func traceID(c *fiber.Ctx) string {
	spanCtx := trace.SpanContextFromContext(c.UserContext())
	if !spanCtx.HasTraceID() {
		return ""
	}

	return spanCtx.TraceID().String()
}
