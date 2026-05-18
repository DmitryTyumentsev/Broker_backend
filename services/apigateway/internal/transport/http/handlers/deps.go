package handlers

import (
	"Donate_backend/services/apigateway/internal/transport/http/handlers/authhandlers"
)

type Deps struct {
	Auth *authhandlers.AuthHandler
}
