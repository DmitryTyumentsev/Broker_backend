package auth

import (
	"context"
	"strings"

	sharedauth "Broker_backend/shared/pkg/auth"
	"Broker_backend/shared/pkg/requestctx"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	principalUserIDKey   = "principal-user-id"
	principalDeviceIDKey = "principal-device-id"
	principalRoleKey     = "principal-role"
	requestIDKey         = "x-request-id"
)

func InjectOutgoingContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		md = md.Copy()
	} else {
		md = metadata.MD{}
	}

	if principal, ok := sharedauth.PrincipalFromContext(ctx); ok {
		md.Set(principalUserIDKey, principal.UserID)
		md.Set(principalDeviceIDKey, principal.DeviceID)
		md.Set(principalRoleKey, principal.Role)
	}

	if requestID, ok := requestctx.RequestIDFromContext(ctx); ok {
		md.Set(requestIDKey, requestID)
	}

	otel.GetTextMapPropagator().Inject(ctx, metadataCarrier{md: md})

	return metadata.NewOutgoingContext(ctx, md)
}

func RequirePrincipalUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		ctx = ExtractIncomingContext(ctx)

		principal, ok := sharedauth.PrincipalFromContext(ctx)
		if !ok || !principal.Valid() {
			return nil, status.Error(codes.Unauthenticated, "principal is missing")
		}

		return handler(ctx, req)
	}
}

func AuthUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return RequirePrincipalUnaryServerInterceptor()
}

func ExtractIncomingContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	ctx = otel.GetTextMapPropagator().Extract(ctx, metadataCarrier{md: md})

	if requestID := firstMetadataValue(md, requestIDKey); requestID != "" {
		ctx = requestctx.WithRequestID(ctx, requestID)
	}

	principal := sharedauth.Principal{
		UserID:   firstMetadataValue(md, principalUserIDKey),
		DeviceID: firstMetadataValue(md, principalDeviceIDKey),
		Role:     firstMetadataValue(md, principalRoleKey),
	}
	if principal.Valid() {
		ctx = sharedauth.WithPrincipal(ctx, principal)
	}

	return ctx
}

func firstMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}

	return strings.TrimSpace(values[0])
}

type metadataCarrier struct {
	md metadata.MD
}

func (c metadataCarrier) Get(key string) string {
	return firstMetadataValue(c.md, strings.ToLower(key))
}

func (c metadataCarrier) Set(key string, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}

	c.md.Set(strings.ToLower(key), value)
}

func (c metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(c.md))
	for key := range c.md {
		keys = append(keys, key)
	}

	return keys
}
