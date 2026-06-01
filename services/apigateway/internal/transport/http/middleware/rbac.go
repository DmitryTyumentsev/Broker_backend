package middleware

import (
	"strings"

	"Broker_backend/services/apigateway/internal/transport/http/httperr"

	"github.com/gofiber/fiber/v2"
)

func RBAC(allowedRoles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		role = strings.TrimSpace(role)
		if role != "" {
			allowed[role] = struct{}{}
		}
	}

	return func(c *fiber.Ctx) error {
		claims, ok := CurrentClaims(c)
		if !ok {
			return httperr.WriteUnauthorized(c, "auth context is missing")
		}

		if len(allowed) == 0 {
			return httperr.WriteForbidden(c, "no roles are allowed")
		}

		if _, ok := allowed[claims.Role]; !ok {
			return httperr.WriteForbidden(c, "insufficient role")
		}

		return c.Next()
	}
}
