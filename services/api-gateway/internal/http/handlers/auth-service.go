package handlers

import (
	"Donate_backend/services/api-gateway/internal/http/dto"
	"Donate_backend/services/api-gateway/internal/http/errors"

	"github.com/gofiber/fiber/v3"
)

func (h *AuthService) Register(c fiber.Ctx) error {
	var req dto.RegisterUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(errors.ErrWrongCredentials.Error())
	}

	return nil
}
