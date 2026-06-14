package http

import (
	"Broker_backend/services/apigateway/internal/transport/http/middleware"

	"github.com/gofiber/fiber/v2"
)

func postJSON[T any](
	router fiber.Router,
	path string,
	validator middleware.RequestValidator,
	handler fiber.Handler,
) fiber.Router {
	return jsonRoute[T](router, fiber.MethodPost, path, validator, handler)
}

func jsonRoute[T any](
	router fiber.Router,
	method string,
	path string,
	validator middleware.RequestValidator,
	handler fiber.Handler,
) fiber.Router {
	return router.Add(method, path, middleware.ValidateJSON[T](validator), handler)
}
