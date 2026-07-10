package sharedauth

import (
	"context"
	"testing"

	sharedauth "Broker_backend/shared/pkg/auth"
	"Broker_backend/shared/pkg/requestctx"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestInjectAndExtractPrincipalMetadata(t *testing.T) {
	ctx := context.Background()
	ctx = sharedauth.WithPrincipal(ctx, sharedauth.Principal{
		UserID:   "user-1",
		DeviceID: "device-1",
		Role:     "broker_team_member",
	})
	ctx = requestctx.WithRequestID(ctx, "request-1")

	outgoing := InjectOutgoingContext(ctx)
	md, ok := metadata.FromOutgoingContext(outgoing)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}

	incoming := metadata.NewIncomingContext(context.Background(), md)
	extracted := ExtractIncomingContext(incoming)

	principal, ok := sharedauth.PrincipalFromContext(extracted)
	if !ok {
		t.Fatal("expected principal in context")
	}

	if principal.UserID != "user-1" || principal.DeviceID != "device-1" || principal.Role != "broker_team_member" {
		t.Fatalf("unexpected principal: %+v", principal)
	}

	requestID, ok := requestctx.RequestIDFromContext(extracted)
	if !ok || requestID != "request-1" {
		t.Fatalf("unexpected request id: %q", requestID)
	}
}

func TestAuthUnaryServerInterceptorRejectsMissingPrincipal(t *testing.T) {
	interceptor := AuthUnaryServerInterceptor()

	_, err := interceptor(
		context.Background(),
		struct{}{},
		&grpc.UnaryServerInfo{FullMethod: "/partner.v1.PartnerService/Create"},
		func(ctx context.Context, req any) (any, error) {
			t.Fatal("handler must not be called")
			return nil, nil
		},
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %s", status.Code(err))
	}
}

func TestAuthUnaryServerInterceptorAllowsPrincipal(t *testing.T) {
	md := metadata.Pairs(
		principalUserIDKey, "user-1",
		principalDeviceIDKey, "device-1",
		principalRoleKey, "broker_team_member",
	)
	ctx := metadata.NewIncomingContext(context.Background(), md)
	interceptor := AuthUnaryServerInterceptor()

	called := false
	_, err := interceptor(
		ctx,
		struct{}{},
		&grpc.UnaryServerInfo{FullMethod: "/partner.v1.PartnerService/Create"},
		func(ctx context.Context, req any) (any, error) {
			called = true
			if _, ok := sharedauth.PrincipalFromContext(ctx); !ok {
				t.Fatal("expected principal in handler context")
			}
			return nil, nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Fatal("expected handler call")
	}
}
