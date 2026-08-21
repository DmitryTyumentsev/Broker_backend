package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"

	"Broker_backend/shared/pkg/requestctx"

	"github.com/gofiber/fiber/v2"
)

const (
	RequestIDHeader   = "X-Request-ID"
	requestIDLocalKey = "request.id"
	maxRequestIDLen   = 128
)

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := normalizeRequestID(c.Get(RequestIDHeader))
		if requestID == "" {
			requestID = newRequestID()
		}

		c.Set(RequestIDHeader, requestID)
		c.Locals(requestIDLocalKey, requestID)
		ctx := requestctx.WithRequestID(userContext(c), requestID)
		c.SetUserContext(requestctx.WithClientIP(ctx, c.IP()))

		return c.Next()
	}
}

func CurrentRequestID(c *fiber.Ctx) string {
	if c == nil {
		return ""
	}

	if requestID, ok := c.Locals(requestIDLocalKey).(string); ok {
		return requestID
	}

	requestID, _ := requestctx.RequestIDFromContext(c.UserContext())
	return requestID
}

func normalizeRequestID(requestID string) string {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" || len(requestID) > maxRequestIDLen {
		return ""
	}

	for _, r := range requestID {
		if r < 33 || r > 126 {
			return ""
		}
	}

	return requestID
}

func newRequestID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return ""
	}

	return hex.EncodeToString(raw[:])
}

func userContext(c *fiber.Ctx) context.Context {
	ctx := c.UserContext()
	if ctx == nil {
		return context.Background()
	}

	return ctx
}
