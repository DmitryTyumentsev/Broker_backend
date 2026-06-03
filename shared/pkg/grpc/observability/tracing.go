package observability

import (
	"context"
	"strings"

	grpcauth "Broker_backend/shared/pkg/grpc/auth"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func TraceUnaryServerInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		serviceName = "grpc-service"
	}

	tracer := otel.Tracer(serviceName)

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		ctx = grpcauth.ExtractIncomingContext(ctx)
		ctx, span := tracer.Start(
			ctx,
			info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(attribute.String("rpc.method", info.FullMethod)),
		)
		defer span.End()

		resp, err := handler(ctx, req)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, status.Code(err).String())
		}

		return resp, err
	}
}
