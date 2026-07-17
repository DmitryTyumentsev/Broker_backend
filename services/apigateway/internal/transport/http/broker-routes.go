package http

import (
	"Broker_backend/services/apigateway/internal/transport/http/dto/brokerdto"
	"Broker_backend/services/apigateway/internal/transport/http/handlers"
	"Broker_backend/services/apigateway/internal/transport/http/middleware"
	"Broker_backend/shared/pkg/authz/permissions"

	"github.com/gofiber/fiber/v2"
)

func registerBrokerRoutes(r fiber.Router, h *handlers.Deps) {
	brokers := r.Group("/brokers")

	registerCustomerFixationRoutes(brokers, h)
}

func registerCustomerFixationRoutes(brokers fiber.Router, h *handlers.Deps) {
	customerFixations := brokers.Group("/customer-fixation")

	createCustomerFixation(customerFixations, h)
	readCustomerFixation(customerFixations, h)
	updateCustomerFixation(customerFixations, h)
}

func createCustomerFixation(r fiber.Router, h *handlers.Deps) {
	postJSON[brokerdto.FixationCustomerRequest](
		r,
		"/",
		h.Validator,
		[]fiber.Handler{
			middleware.RequirePermission(h.Authz, permissions.CustomerFixationCreate),
		},
		h.Broker.CreateFixationCustomer,
	)
}

func readCustomerFixation(r fiber.Router, h *handlers.Deps) {
	get(
		"/:customer_id",
		[]fiber.Handler{
			middleware.RequirePermission(h.Authz, permissions.CustomerFixationRead),
		},
		h.Broker.ReadFixationCustomer,
	)
}

func updateCustomerFixation(r fiber.Router, h *handlers.Deps) {
	patchJSON[brokerdto.FixationCustomerRequest](
		r,
		"/",
		h.Validator,
		[]fiber.Handler{
			middleware.RequirePermission(h.Authz, permissions.CustomerFixationUpdate),
		},
		h.Broker.UpdateFixationCustomer,
	)
}
