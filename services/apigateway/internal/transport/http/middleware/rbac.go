package middleware

import (
	"strings"

	"Broker_backend/services/apigateway/internal/transport/http/httperr"

	"github.com/gofiber/fiber/v2"
)

func RBAC(allowedRoles ...string) fiber.Handler { // ...string и []string одно и тоже? если нет, зачем ставят многоточие перед типом?
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
		} //немного не понимаю что с чем мы сравниваем и откуда берем - мы получаем в fiber.Ctx данные, кладем их в localKey. Затем вытаскиваем оттуда access токен пользователя. как мы его проверяем? где? что в нем? и второй важный момент - а где само сравнение ролей? что с чем сравниваем? я читал код и не понял ничего, показывай прям на нем что где

		return c.Next()
	}
}
