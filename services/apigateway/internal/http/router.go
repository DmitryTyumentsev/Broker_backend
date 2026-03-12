package http

import (
	"Donate_backend/services/apigateway/internal/http/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRouter(app *fiber.App, h *handlers.Deps) {
	api := app.Group("/api")

	//=====GLOBAL MIDDLEWARES=====

	//=====api/v1=====
	v1 := api.Group("/v1")
	v1.Post("/register", h.Auth)
	v1.Post("/login", h.Login)
	v1.Post("/refresh", h.Refresh)
	v1.Post("/logout", h.Logout)

}
