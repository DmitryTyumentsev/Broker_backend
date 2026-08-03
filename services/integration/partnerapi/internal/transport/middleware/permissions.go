package middleware

import (
	"Broker_backend/services/integration/partnerapi/internal/transport/grpc/grpcerr"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthzPolicy interface {
	IsAllowed(role string, permission string) bool
}

func RequirePermission(policy AuthzPolicy, permission string) fiber.Handler {
	permission = strings.TrimSpace(permission)

	return func(c *fiber.Ctx) error {
		if policy == nil {
			return grpcerr.WriteServiceUnavailable(c, "authorization policy is not configured")
		}

		if permission == "" {
			return grpcerr.WriteForbidden(c, "permission is not configured")
		}

		principal, ok := CurrentPrincipal(c)
		if !ok {
			return grpcerr.WriteUnauthorized(c, "auth context is missing")
		}

		if !policy.IsAllowed(principal.Role, permission) {
			return grpcerr.WriteForbidden(c, "insufficient permissions")
		}

		return c.Next()
	}
}
