package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func TestRecoveryConvertsPanicToInternalServerError(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Use(Recovery(zap.NewNop()))
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("boom")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/panic", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", resp.StatusCode)
	}
}
