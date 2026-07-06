package http

import (
	"Broker_backend/services/apigateway/internal/transport/http/dto/brokerdto"
	"Broker_backend/services/apigateway/internal/transport/http/handlers"
	"Broker_backend/services/apigateway/internal/transport/http/middleware"
	"Broker_backend/shared/pkg/authz/permissions"

	"github.com/gofiber/fiber/v2"
)

func registerBrokerRoutes(r fiber.Router, h *handlers.Deps) {
	broker := r.Group(
		"/brokers",
		middleware.RequirePermission(h.Authz, permissions.CustomerFixationCreate),
	)
	postJSON[brokerdto.ConnectCustomerRequest](broker, "/connect-customer", h.Validator, h.Broker.ConnectCustomer)
}
