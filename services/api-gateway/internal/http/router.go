package http

import (
	"Donate_backend/services/api-gateway/internal/http/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupRouter(h *handlers.Handlers) {
	app := fiber.New()
	r := app.Group("/api")

	//=====GLOBAL MIDDLEWARES=====

	//=====api/v1=====
	r.Group("/v1")
	r.Group("/register", h.Auth.Register)
	r.Group("/login", h.Auth.Login)
	r.Group("/refresh", h.Auth.Refresh)
	r.Group("/logout", h.Auth.Logout)

}
