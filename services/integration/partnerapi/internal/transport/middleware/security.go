package middleware

import (
	"Broker_backend/services/integration/partnerapi/internal/config"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func SecurityHeaders(enabled bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if enabled {
			c.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			c.Set("X-Content-Type-Options", "nosniff")
			c.Set("X-Frame-Options", "DENY")
			c.Set("X-XSS-Protection", "0")
			c.Set("Referrer-Policy", "no-referrer")
			c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			c.Set("Cross-Origin-Opener-Policy", "same-origin")
		}

		return c.Next()
	}
}

func CORS(cfg config.CORSConfig) fiber.Handler {
	allowedOrigins := stringSet(cfg.AllowOrigins)
	allowedMethods := strings.Join(nonEmptyStrings(cfg.AllowMethods), ", ")
	allowedHeaders := strings.Join(nonEmptyStrings(cfg.AllowHeaders), ", ")
	exposeHeaders := strings.Join(nonEmptyStrings(cfg.ExposeHeaders), ", ")

	return func(c *fiber.Ctx) error {
		if !cfg.Enabled {
			return c.Next()
		}

		origin := strings.TrimSpace(c.Get(fiber.HeaderOrigin))
		if origin != "" && originAllowed(origin, allowedOrigins) {
			if cfg.AllowCredentials && allowedOrigins["*"] {
				c.Set(fiber.HeaderAccessControlAllowOrigin, origin)
			} else if allowedOrigins["*"] {
				c.Set(fiber.HeaderAccessControlAllowOrigin, "*")
			} else {
				c.Set(fiber.HeaderAccessControlAllowOrigin, origin)
			}

			if cfg.AllowCredentials {
				c.Set(fiber.HeaderAccessControlAllowCredentials, "true")
			}

			if allowedMethods != "" {
				c.Set(fiber.HeaderAccessControlAllowMethods, allowedMethods)
			}

			if allowedHeaders != "" {
				c.Set(fiber.HeaderAccessControlAllowHeaders, allowedHeaders)
			}

			if exposeHeaders != "" {
				c.Set(fiber.HeaderAccessControlExposeHeaders, exposeHeaders)
			}

			if cfg.MaxAgeSeconds > 0 {
				c.Set(fiber.HeaderAccessControlMaxAge, strconv.Itoa(cfg.MaxAgeSeconds))
			}
		}

		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return c.Next()
	}
}

func originAllowed(origin string, allowed map[string]bool) bool {
	return allowed["*"] || allowed[origin]
}

func stringSet(values []string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out[value] = true
		}
	}

	return out
}

func nonEmptyStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}

	return out
}
