package grpc

import (
	"Broker_backend/services/app/authservice/internal/domain"
	"Broker_backend/services/app/authservice/internal/usecase"
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapErrorSanitizesWrappedErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		code    codes.Code
		message string
	}{
		{
			name:    "validation error",
			err:     fmt.Errorf("usecase.Register: %w", usecase.ErrEmailInvalid),
			code:    codes.InvalidArgument,
			message: usecase.ErrEmailInvalid.Error(),
		},
		{
			name:    "unique email error",
			err:     fmt.Errorf("usecase.Register: save user: postgres.users.Save: %w", domain.ErrNotUniqueEmail),
			code:    codes.AlreadyExists,
			message: "email already exists",
		},
		{
			name:    "auth error",
			err:     fmt.Errorf("usecase.Login: %w", domain.ErrUnauthenticated),
			code:    codes.Unauthenticated,
			message: "invalid credentials or token",
		},
		{
			name:    "unknown error",
			err:     errors.New("database connection leaked detail"),
			code:    codes.Internal,
			message: "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, ok := status.FromError(mapError(tt.err))
			if !ok {
				t.Fatal("expected grpc status error")
			}

			if st.Code() != tt.code {
				t.Fatalf("expected code %s, got %s", tt.code, st.Code())
			}

			if st.Message() != tt.message {
				t.Fatalf("expected message %q, got %q", tt.message, st.Message())
			}
		})
	}
}
