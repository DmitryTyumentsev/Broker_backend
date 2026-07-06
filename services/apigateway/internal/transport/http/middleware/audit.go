package middleware

import (
	sharedauth "Broker_backend/shared/pkg/authz"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func AuditLog(c *fiber.Ctx, logger *zap.Logger, event string, fields ...zap.Field) {
	if logger == nil {
		logger = zap.NewNop()
	}

	baseFields := []zap.Field{
		zap.String("event", event),
		zap.String("request_id", CurrentRequestID(c)),
		zap.String("ip", c.IP()),
		zap.String("user_agent", c.Get(fiber.HeaderUserAgent)),
	}

	if principal, ok := sharedauth.PrincipalFromContext(c.UserContext()); ok {
		baseFields = append(
			baseFields,
			zap.String("user_id", principal.UserID),
			zap.String("device_id", principal.DeviceID),
			zap.String("role", principal.Role),
		)
	}

	logger.Info("security audit", append(baseFields, fields...)...)
}
