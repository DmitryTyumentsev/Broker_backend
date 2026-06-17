package http

import (
	"Broker_backend/services/apigateway/internal/transport/http/dto/brokerdto"
	"Broker_backend/services/apigateway/internal/transport/http/handlers"

	"github.com/gofiber/fiber/v2"
)

func registerBrokerRoutes(r fiber.Router, h *handlers.Deps) {
	postJSON[brokerdto.ConnectCustomerRequest](r, "/connect-customer", h.Validator, h.Broker.ConnectCustomer)

}
