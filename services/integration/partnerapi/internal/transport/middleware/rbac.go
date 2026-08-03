package middleware

import (
	"Broker_backend/services/integration/partnerapi/internal/transport/grpc/grpcerr"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func RequireRole(allowedRoles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(allowedRoles))

	for _, role := range allowedRoles {
		role = strings.TrimSpace(role)
		if role != "" {
			allowed[role] = struct{}{}
		}
	}

	return func(c *fiber.Ctx) error {
		principal, ok := CurrentPrincipal(c)
		if !ok {
			return grpcerr.WriteUnauthorized(c, "auth context is missing")
		}

		if len(allowed) == 0 {
			return grpcerr.WriteForbidden(c, "no roles are allowed")
		}

		if _, ok := allowed[principal.Role]; !ok {
			return grpcerr.WriteForbidden(c, "insufficient role")
		}

		return c.Next()
	}
}

func RBAC(allowedRoles ...string) fiber.Handler {
	return RequireRole(allowedRoles...)
}
