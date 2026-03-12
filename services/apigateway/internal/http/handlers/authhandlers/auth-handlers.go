package authhandlers

import (
	"Donate_backend/services/apigateway/internal/clients/authclient"
	"Donate_backend/services/apigateway/internal/http/dto"
	"Donate_backend/services/apigateway/internal/http/errors"
	"Donate_backend/services/authservice/internal/config"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	config     *config.Config
	logger     *zap.Logger
	authclient authclient.Handlers
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors.ErrWrongCredentials.Error())
	}

	return nil
}
