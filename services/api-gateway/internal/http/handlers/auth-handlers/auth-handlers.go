package auth_handlers

import (
	"Donate_backend/services/api-gateway/internal/http/dto"
	"Donate_backend/services/api-gateway/internal/http/errors"
	"Donate_backend/services/api-gateway/internal/http/handlers"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	h *handlers.Handlers //TODO: почему я не могу указать в ресивере ниже только поле h, а не тянуть всю структуру?
}

func (h *AuthHandler) Register(c *fiber.Ctx) error { //TODO: почему я могу h *handlers.Handlers передать на вход, но если передаю как ресивер компилятор просит объявить в этом пакете структуру? не понимаю зачем это правило
	var req dto.RegisterUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errors.ErrWrongCredentials.Error())
	}

	return nil
}
