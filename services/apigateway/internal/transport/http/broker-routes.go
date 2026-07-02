package http

import (
	"Broker_backend/services/apigateway/internal/transport/http/dto/brokerdto"
	"Broker_backend/services/apigateway/internal/transport/http/handlers"
	"Broker_backend/services/apigateway/internal/transport/http/middleware"

	"github.com/gofiber/fiber/v2"
)

func registerBrokerRoutes(r fiber.Router, h *handlers.Deps) {
	adminsAndTeamLeads := r.Group(
		"/brokers",
		middleware.RBAC(h.Config.Business.BrokerAdminAllowedRoles..., h.Config.Business.BrokerTeamLeadAllowedRoles...,),
	) //как правильно должно быть? зачем три точки ставим? логика что крепить могут только тимлиды и админы
	postJSON[brokerdto.ConnectCustomerRequest](adminsAndTeamLeads, "/connect-customer", h.Validator, h.Broker.ConnectCustomer)
}
