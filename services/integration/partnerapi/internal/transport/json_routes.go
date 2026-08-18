package transport

import (
	"Broker_backend/services/integration/partnerapi/internal/transport/middleware"

	"github.com/gofiber/fiber/v2"
)

func postJSON[T any]( //добавь миддлвар на get
	router fiber.Router,
	path string,
	validator middleware.RequestValidator,
	routeMiddlewares []fiber.Handler,
	handler fiber.Handler,
) fiber.Router {
	return jsonRoute[T](
		router,
		fiber.MethodPost,
		path,
		validator,
		routeMiddlewares,
		handler,
	)
}

func putJSON[T any](
	router fiber.Router,
	path string,
	validator middleware.RequestValidator,
	routeMiddlewares []fiber.Handler,
	handler fiber.Handler,
) fiber.Router {
	return jsonRoute[T](
		router,
		fiber.MethodPut,
		path,
		validator,
		routeMiddlewares,
		handler,
	)
}

func patchJSON[T any](
	router fiber.Router,
	path string,
	validator middleware.RequestValidator,
	routeMiddlewares []fiber.Handler,
	handler fiber.Handler,
) fiber.Router {
	return jsonRoute[T](
		router,
		fiber.MethodPatch,
		path,
		validator,
		routeMiddlewares,
		handler,
	)
}

func jsonRoute[T any](
	router fiber.Router,
	method string,
	path string,
	validator middleware.RequestValidator,
	routeMiddlewares []fiber.Handler,
	handler fiber.Handler,
) fiber.Router {
	chain := make([]fiber.Handler, 0, len(routeMiddlewares)+2)
	chain = append(chain, routeMiddlewares...)
	chain = append(chain, middleware.ValidateJSON[T](validator))
	chain = append(chain, handler)

	return router.Add(method, path, chain...)
}
