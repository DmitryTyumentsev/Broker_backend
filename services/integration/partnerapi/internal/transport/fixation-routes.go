package transport

import (
	"Broker_backend/services/integration/partnerapi/internal/transport/dto/fixationdto"
	"Broker_backend/services/integration/partnerapi/internal/transport/handlers"
	"Broker_backend/services/integration/partnerapi/internal/transport/middleware"
	"Broker_backend/shared/pkg/authz/permissions"

	"github.com/gofiber/fiber/v2"
)

func registerFixationRoutes(brokers fiber.Router, h *handlers.Deps) {
	fixations := brokers.Group("/fixations")

	newFixation(fixations, h)
}

func newFixation(r fiber.Router, h *handlers.Deps) {
	postJSON[fixationdto.FixationRequest](
		r,
		"/",
		h.Validator,
		[]fiber.Handler{
			middleware.RequirePermission(h.Authz, permissions.FixationNew),
		},
		h.Fixation.NewFixation,
	)
}
