package http

import (
	"Broker_backend/services/apigateway/internal/transport/http/dto/authdto"
	"Broker_backend/services/apigateway/internal/transport/http/handlers"
	"Broker_backend/services/apigateway/internal/transport/http/middleware"
	"Broker_backend/shared/pkg/authz/permissions"

	"github.com/gofiber/fiber/v2"
)

func registerAuthRoutes(v1 fiber.Router, h *handlers.Deps) {
	auth := v1.Group("/auth")
	auth.Use(middleware.RedisRateLimit(h.Redis, h.Config.Business.AuthRateLimit, h.Logger))

	postJSON[authdto.RegisterRequest](auth, "/register", h.Validator, nil, h.Auth.Register)
	postJSON[authdto.LoginRequest](auth, "/login", h.Validator, nil, h.Auth.Login)
	postJSON[authdto.RefreshRequest](auth, "/refresh", h.Validator, nil, h.Auth.Refresh)
	postJSON[authdto.LogoutRequest](auth, "/logout", h.Validator, nil, h.Auth.Logout)
}

func registerAuthProtectedRoutes(protected fiber.Router, h *handlers.Deps) {
	protected.Get("/me", h.Auth.Me)

	protected.Get(
		"/admin/ping",
		middleware.RequirePermission(h.Authz, permissions.DeveloperAdminAccess),
		h.Auth.AdminPing,
	)
}
