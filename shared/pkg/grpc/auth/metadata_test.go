package auth

import (
	"context"
	"testing"

	"Broker_backend/shared/pkg/authz"
	"Broker_backend/shared/pkg/requestctx"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	testAgencyID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testUserID   = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

func TestInjectAndExtractPrincipalMetadata(t *testing.T) {
	ctx := context.Background()
	ctx = authz.WithPrincipal(ctx, authz.Principal{
		AgencyID: testAgencyID,
		UserID:   testUserID,
		DeviceID: "device-1",
		Role:     "broker_team_member",
	})
	ctx = requestctx.WithRequestID(ctx, "request-1")
	ctx = requestctx.WithClientIP(ctx, "192.0.2.10")

	outgoing := InjectOutgoingContext(ctx)
	md, ok := metadata.FromOutgoingContext(outgoing)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}

	incoming := metadata.NewIncomingContext(context.Background(), md)
	extracted := ExtractIncomingContext(incoming)

	principal, ok := authz.PrincipalFromContext(extracted)
	if !ok {
		t.Fatal("expected principal in context")
	}

	if principal.AgencyID != testAgencyID || principal.UserID != testUserID ||
		principal.DeviceID != "device-1" || principal.Role != "broker_team_member" {
		t.Fatalf("unexpected principal: %+v", principal)
	}

	requestID, ok := requestctx.RequestIDFromContext(extracted)
	if !ok || requestID != "request-1" {
		t.Fatalf("unexpected request id: %q", requestID)
	}

	clientIP, ok := requestctx.ClientIPFromContext(extracted)
	if !ok || clientIP != "192.0.2.10" {
		t.Fatalf("unexpected client ip: %q", clientIP)
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
		principalAgencyIDKey, testAgencyID.String(),
		principalUserIDKey, testUserID.String(),
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
			if _, ok := authz.PrincipalFromContext(ctx); !ok {
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
