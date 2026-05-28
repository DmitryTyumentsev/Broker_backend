package http

import (
	"Broker_backend/services/apigateway/internal/transport/http/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRouter(app *fiber.App, h *handlers.Deps) *fiber.App {
	api := app.Group("/api")

	//=====GLOBAL MIDDLEWARES=====

	//=====api/v1=====
	v1 := api.Group("/v1")

	//=====authservice=====
	auth := v1.Group("/auth")
	auth.Post("/register", h.Auth.Register)
	auth.Post("/login", h.Auth.Login)
	auth.Post("/refresh", h.Auth.Refresh)
	auth.Post("/logout", h.Auth.Logout)

	return app
}
