package authz

import (
	"context"
	"strings"

	sharedjwt "Broker_backend/shared/pkg/security/jwt"

	"github.com/google/uuid"
)

type contextKey string

const principalContextKey contextKey = "principal"

type Principal struct {
	AgencyID uuid.UUID
	UserID   uuid.UUID
	DeviceID string
	Role     string
}

func PrincipalFromAccessTokenClaims(claims sharedjwt.AccessTokenClaims) Principal {
	return Principal{
		AgencyID: claims.AgencyID,
		UserID:   claims.UserID,
		DeviceID: strings.TrimSpace(claims.DeviceID),
		Role:     strings.TrimSpace(claims.Role),
	}
}

func (p Principal) Valid() bool {
	return p.AgencyID != uuid.Nil &&
		p.UserID != uuid.Nil &&
		strings.TrimSpace(p.DeviceID) != "" &&
		strings.TrimSpace(p.Role) != ""
}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, principalContextKey, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	if ctx == nil {
		return Principal{}, false
	}

	principal, ok := ctx.Value(principalContextKey).(Principal)
	return principal, ok && principal.Valid()
}
