package middleware

import (
	"Broker_backend/services/integration/partnerapi/internal/transport/http/dto"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type validationTestRequest struct {
	Email string `json:"email" validate:"required,email"`
}

func TestValidateJSONRejectsInvalidBody(t *testing.T) {
	app := fiber.New()
	app.Post("/validate", ValidateJSON[validationTestRequest](NewRequestValidator()), func(c *fiber.Ctx) error {
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

	var body dto.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if len(body.Fields) != 1 {
		t.Fatalf("expected one validation field, got %d", len(body.Fields))
	}

	if body.Fields[0].Field != "email" {
		t.Fatalf("expected json field name email, got %q", body.Fields[0].Field)
	}
}

func TestValidateJSONStoresValidatedBody(t *testing.T) {
	app := fiber.New()
	app.Post("/validate", ValidateJSON[validationTestRequest](NewRequestValidator()), func(c *fiber.Ctx) error {
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
