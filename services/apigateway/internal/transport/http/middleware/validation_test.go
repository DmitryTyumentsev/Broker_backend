package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	validate "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type validationTestRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func TestValidateJSONRejectsInvalidBody(t *testing.T) {
	app := fiber.New()
	app.Post("/validate", ValidateJSON[validationTestRequest](validate.New()), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(fiber.MethodPost, "/validate", strings.NewReader(`{"email":"bad"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestValidateJSONStoresValidatedBody(t *testing.T) {
	app := fiber.New()
	app.Post("/validate", ValidateJSON[validationTestRequest](validate.New()), func(c *fiber.Ctx) error {
		req, ok := ValidatedBody[validationTestRequest](c)
		if !ok {
			t.Fatal("expected validated body")
		}

		if req.Email != "user@example.com" {
			t.Fatalf("unexpected email: %s", req.Email)
		}

		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(fiber.MethodPost, "/validate", strings.NewReader(`{"email":"user@example.com"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}
