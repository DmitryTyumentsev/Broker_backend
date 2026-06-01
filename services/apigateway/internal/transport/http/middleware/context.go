package middleware

import (
	"context"
	"time"

	"Broker_backend/services/apigateway/internal/transport/http/httperr"

	"github.com/gofiber/fiber/v2"
)

func RequestTimeout(timeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if timeout <= 0 {
			return c.Next()
		}

		parent := c.UserContext()
		if parent == nil {
			parent = context.Background()
		}

		ctx, cancel := context.WithTimeout(parent, timeout)
		defer cancel()

		c.SetUserContext(ctx)

		if err := c.Next(); err != nil {
			return err
		}

		if ctx.Err() == context.DeadlineExceeded {
			return httperr.WriteServiceUnavailable(c, "request deadline exceeded")
		}

		return nil
	}
}
