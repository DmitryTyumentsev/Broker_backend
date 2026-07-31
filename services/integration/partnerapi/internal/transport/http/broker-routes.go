package http

import (
	"Broker_backend/services/integration/partnerapi/internal/transport/http/dto/brokerdto"
	"Broker_backend/services/integration/partnerapi/internal/transport/http/handlers"
	"Broker_backend/services/integration/partnerapi/internal/transport/http/middleware"
	"Broker_backend/shared/pkg/authz/permissions"

	"github.com/gofiber/fiber/v2"
)

func registerBrokerRoutes(r fiber.Router, h *handlers.Deps) {
	brokers := r.Group("/brokers")

	registerFixationRoutes(brokers, h)
}

func registerFixationRoutes(brokers fiber.Router, h *handlers.Deps) {
	fixations := brokers.Group("/fixations")

	newFixation(fixations, h)
}

func newFixation(r fiber.Router, h *handlers.Deps) {
	postJSON[brokerdto.FixationRequest](
		r,
		"/",
		h.Validator,
		[]fiber.Handler{
			middleware.RequirePermission(h.Authz, permissions.FixationNew),
		},
		h.Broker.NewFixation,
	)
}
