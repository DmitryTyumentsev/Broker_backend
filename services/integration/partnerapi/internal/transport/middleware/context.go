package middleware

import (
	"Broker_backend/services/integration/partnerapi/internal/transport/grpc/grpcerr"
	"context"
	"time"

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

		if ctx.Err() == context.DeadlineExceeded && c.Response().StatusCode() < fiber.StatusBadRequest {
			return grpcerr.WriteGatewayTimeout(c, "request deadline exceeded")
		}

		return nil
	}
}
