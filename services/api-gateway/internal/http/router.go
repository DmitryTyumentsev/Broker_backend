package http

import (
	"Donate_backend/services/api-gateway/internal/http/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRouter(h *handlers.Handlers) {
	app := fiber.New()
	app.Group("/api")
	ctx := new(fiber.Ctx)

	//=====GLOBAL MIDDLEWARES=====

	//=====api/v1=====
	app.Group("/v1")
	app.Group("/register", h.Auth.Register(ctx))
	app.Group("/login", h.Auth.Login)
	app.Group("/refresh", h.Auth.Refresh)
	app.Group("/logout", h.Auth.Logout)

}
