package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequestIDUsesExistingHeader(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/ping", func(c *fiber.Ctx) error {
		if CurrentRequestID(c) != "request-1" {
			t.Fatalf("unexpected request id: %q", CurrentRequestID(c))
		}

		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/ping", nil)
	req.Header.Set(RequestIDHeader, "request-1")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if got := resp.Header.Get(RequestIDHeader); got != "request-1" {
		t.Fatalf("expected request id header, got %q", got)
	}
}
