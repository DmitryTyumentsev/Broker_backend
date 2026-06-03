package middleware

import (
	"errors"

	"Broker_backend/services/apigateway/internal/transport/http/dto"

	validate "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

const validatedBodyLocalKey = "request.validated_body"

type RequestValidator interface {
	Struct(s any) error
}

func ValidateJSON[T any](validator RequestValidator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req T

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Code:    fiber.StatusBadRequest,
				Message: "invalid request body",
			})
		}

		if validator != nil {
			if err := validator.Struct(req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
					Code:    fiber.StatusBadRequest,
					Message: "validation failed",
					Fields:  validationFields(err),
				})
			}
		}

		c.Locals(validatedBodyLocalKey, req)
		return c.Next()
	}
}

func ValidatedBody[T any](c *fiber.Ctx) (T, bool) {
	if c == nil {
		var zero T
		return zero, false
	}

	req, ok := c.Locals(validatedBodyLocalKey).(T)
	return req, ok
}

func validationFields(err error) []dto.Field {
	var validationErrors validate.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return nil
	}

	fields := make([]dto.Field, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		fields = append(fields, dto.Field{
			Field:   fieldErr.Field(),
			Message: fieldErr.Tag(),
		})
	}

	return fields
}
