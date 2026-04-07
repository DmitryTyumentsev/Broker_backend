package handlers

import (
	"Donate_backend/services/apigateway/internal/http/handlers/authhandlers"
)

type Deps struct {
	Auth *authhandlers.AuthHandler
}
