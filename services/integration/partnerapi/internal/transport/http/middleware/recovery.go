package middleware

import (
	"Broker_backend/services/integration/partnerapi/internal/transport/http/httperr"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) fiber.Handler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error(
					"http panic recovered",
					zap.Any("panic", recovered),
					zap.String("request_id", CurrentRequestID(c)),
					zap.String("method", c.Method()),
					zap.String("path", c.Path()),
					zap.ByteString("stack", debug.Stack()),
				)
				err = httperr.WriteInternal(c)
			}
		}()

		return c.Next()
	}
}
