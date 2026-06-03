package tracing

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Enabled      bool
	ServiceName  string
	OTLPEndpoint string
	Insecure     bool
	SampleRatio  float64
}

func InitTracerProvider(ctx context.Context, cfg Config) (*sdktrace.TracerProvider, error) {
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	if !cfg.Enabled {
		return sdktrace.NewTracerProvider(), nil
	}

	endpoint := strings.TrimSpace(cfg.OTLPEndpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("otlp endpoint is required")
	}

	options := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
	}
	if cfg.Insecure {
		options = append(options, otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	exporter, err := otlptracegrpc.New(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("create otlp trace exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))),
		sdktrace.WithResource(resource.NewWithAttributes(
			"",
			attribute.String("service.name", strings.TrimSpace(cfg.ServiceName)),
		)),
	)

	otel.SetTracerProvider(tracerProvider)

	return tracerProvider, nil
}
