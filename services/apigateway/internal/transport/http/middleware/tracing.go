package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Trace(serviceName string) fiber.Handler {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		serviceName = "apigateway"
	}

	tracer := otel.Tracer(serviceName)

	return func(c *fiber.Ctx) error {
		ctx := otel.GetTextMapPropagator().Extract(userContext(c), fiberHeaderCarrier{ctx: c})

		spanName := c.Method() + " " + c.Path()
		ctx, span := tracer.Start(
			ctx,
			spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.request.method", c.Method()),
				attribute.String("url.path", c.Path()),
				attribute.String("client.address", c.IP()),
				attribute.String("request.id", CurrentRequestID(c)),
			),
		)
		defer span.End()

		c.SetUserContext(ctx)

		err := c.Next()
		statusCode := c.Response().StatusCode()
		if err != nil && statusCode < fiber.StatusBadRequest {
			statusCode = fiber.StatusInternalServerError
		}

		span.SetAttributes(attribute.Int("http.response.status_code", statusCode))
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, err.Error())
		} else if statusCode >= fiber.StatusInternalServerError {
			span.SetStatus(otelcodes.Error, "server error")
		}

		return err
	}
}

type fiberHeaderCarrier struct {
	ctx *fiber.Ctx
}

func (c fiberHeaderCarrier) Get(key string) string {
	return c.ctx.Get(key)
}

func (c fiberHeaderCarrier) Set(key string, value string) {
	c.ctx.Set(key, value)
}

func (c fiberHeaderCarrier) Keys() []string {
	keys := make([]string, 0)
	c.ctx.Request().Header.VisitAll(func(key []byte, value []byte) {
		keys = append(keys, string(key))
	})

	return keys
}
