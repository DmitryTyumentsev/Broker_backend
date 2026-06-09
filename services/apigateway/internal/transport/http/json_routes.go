package http

import (
	"Broker_backend/services/apigateway/internal/transport/http/middleware"

	"github.com/gofiber/fiber/v2"
)

// validatedJSONRoute keeps JSON body parsing and validation attached to route registration.
func validatedJSONRoute[T any](
	router fiber.Router,
	method string,
	path string,
	validator middleware.RequestValidator,
	handler fiber.Handler,
) fiber.Router {
	return router.Add(method, path, middleware.ValidateJSON[T](validator), handler)
}
