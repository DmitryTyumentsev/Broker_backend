package http

import (
	"Donate_backend/services/api-gateway/internal/http/handlers"
	auth_handlers "Donate_backend/services/api-gateway/internal/http/handlers/auth-handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRouter(h *handlers.Handlers) {
	app := fiber.New()
	api := app.Group("/api")
	ctx := new(fiber.Ctx)

	//=====GLOBAL MIDDLEWARES=====

	//=====api/v1=====
	v1 := api.Group("/v1")
	v1.Group("/register", auth_handlers.
	v1.Group("/login", h.Login)
	v1.Group("/refresh", h.Refresh)
	v1.Group("/logout", h.Logout)

}
