package http

import (
	"errors"

	"Broker_backend/services/apigateway/internal/transport/http/handlers"
	"Broker_backend/services/apigateway/internal/transport/http/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRouter(app *fiber.App, h *handlers.Deps) error {
	if app == nil {
		return errors.New("fiber app is nil")
	}

	if err := h.Validate(); err != nil {
		return err
	}

	cfg := h.Config

	api := app.Group("/api")

	//=====GLOBAL MIDDLEWARES=====
	api.Use(middleware.RequestTimeout(cfg.Business.ContextTimeout))
	api.Use(middleware.RedisRateLimit(h.Redis, cfg.Business.DefaultRateLimit, h.Logger))

	//=====api/v1=====
	v1 := api.Group("/v1")

	//=====authservice=====
	auth := v1.Group("/auth")
	auth.Use(middleware.RedisRateLimit(h.Redis, cfg.Business.AuthRateLimit, h.Logger))
	auth.Post("/register", h.Auth.Register)
	auth.Post("/login", h.Auth.Login)
	auth.Post("/refresh", h.Auth.Refresh)
	auth.Post("/logout", h.Auth.Logout)

	protected := v1.Group(
		"",
		middleware.Auth(h.AccessVerifier),
		middleware.RBAC(cfg.Business.ProtectedAllowedRoles...),
	)
	protected.Get("/me", h.Auth.Me)
	protected.Get(
		"/admin/ping",
		middleware.RBAC(cfg.Business.AdminAllowedRoles...),
		h.Auth.AdminPing,
	)

	return nil
}
