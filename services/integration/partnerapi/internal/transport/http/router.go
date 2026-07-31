package http

import (
	"Broker_backend/services/integration/partnerapi/internal/transport/http/handlers"
	middleware2 "Broker_backend/services/integration/partnerapi/internal/transport/http/middleware"
	"Broker_backend/shared/pkg/authz/permissions"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func SetupRouter(app *fiber.App, h *handlers.Deps) error {
	if app == nil {
		return errors.New("fiber app is nil")
	}

	if err := h.Validate(); err != nil {
		return err
	}

	registerPlatformMiddleware(app, h)
	registerSystemRoutes(app, h)
	registerAPIRoutes(app, h)

	return nil
}

func registerPlatformMiddleware(app *fiber.App, h *handlers.Deps) {
	cfg := h.Config

	app.Use(middleware2.RequestID())
	app.Use(middleware2.AccessLog(h.Logger))
	app.Use(middleware2.Trace(cfg.Observability.Tracing.ServiceName)) // их делают в реальных проектах? если да - зачем? и в целом еще напомни в чем суть идемпотентности? в том чтобы в базу не записали дважды одну и ту же строку(даже скорее чтоб сервер дубль не обрабатывал)? если так - в случае с patch что происходит? patch же меняет то что есть сейчас, не удаляя остальное, а put полностью все стирает и записывает заново? выходит что оба неидемпотенты, идемпотентен только post?
	app.Use(h.Metrics.Middleware())
	app.Use(middleware2.Recovery(h.Logger))
	app.Use(middleware2.SecurityHeaders(cfg.HTTP.SecurityHeaders.Enabled))
	app.Use(middleware2.CORS(cfg.HTTP.CORS))
}

func registerSystemRoutes(app *fiber.App, h *handlers.Deps) {
	if h.Config.Observability.Metrics.Enabled {
		app.Get(metricPath(h.Config.Observability.Metrics.Path), h.Metrics.Handler())
	}

	app.Get("/healthz", healthHandler)
	app.Get("/readyz", healthHandler)
}

func registerAPIRoutes(app *fiber.App, h *handlers.Deps) {
	cfg := h.Config

	api := app.Group("/api")
	api.Use(middleware2.RequestTimeout(cfg.RequestTimeout()))
	api.Use(middleware2.RedisRateLimit(h.Redis, cfg.Business.DefaultRateLimit, h.Logger))

	v1 := api.Group("/v1")

	registerPublicRoutes(v1, h)
	registerProtectedRoutes(v1, h)
}

func registerPublicRoutes(v1 fiber.Router, h *handlers.Deps) {
	registerAuthRoutes(v1, h)
}

func registerProtectedRoutes(v1 fiber.Router, h *handlers.Deps) {
	cfg := h.Config

	protected := v1.Group(
		"",
		middleware2.Auth(h.AccessVerifier),
		middleware2.RequirePermission(h.Authz, permissions.APIProtectedAccess),
		middleware2.Idempotency(h.Redis, cfg.Business.Idempotency, h.Logger),
	)

	registerAuthProtectedRoutes(protected, h)
	registerBrokerRoutes(protected, h)
}

func metricPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/metrics"
	}

	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}

	return path
}

func healthHandler(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
}
