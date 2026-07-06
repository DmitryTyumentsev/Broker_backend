package middleware

import (
	"context"
	"strings"

	"Broker_backend/services/apigateway/internal/transport/http/httperr"
	sharedauth "Broker_backend/shared/pkg/authz"
	sharedjwt "Broker_backend/shared/pkg/security/jwt"

	"github.com/gofiber/fiber/v2"
)

type contextKey string

const (
	claimsLocalKey = "auth.claims"

	userIDContextKey   contextKey = "auth.user_id"
	deviceIDContextKey contextKey = "auth.device_id"
	roleContextKey     contextKey = "auth.role"
)

type AccessTokenVerifier interface {
	Verify(rawToken string) (sharedjwt.AccessTokenClaims, error)
}

func Auth(verifier AccessTokenVerifier) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if verifier == nil {
			return httperr.WriteServiceUnavailable(c, "auth verifier is not configured")
		}

		token, ok := bearerToken(c.Get(fiber.HeaderAuthorization))
		if !ok {
			return httperr.WriteUnauthorized(c, "missing bearer token")
		}

		claims, err := verifier.Verify(token)
		if err != nil {
			return httperr.WriteUnauthorized(c, "invalid bearer token")
		}

		principal := sharedauth.PrincipalFromAccessTokenClaims(claims)
		if !principal.Valid() {
			return httperr.WriteUnauthorized(c, "invalid bearer token")
		}

		c.Locals(claimsLocalKey, claims)
		c.SetUserContext(sharedauth.WithPrincipal(contextWithClaims(c.UserContext(), claims), principal))

		return c.Next()
	}
}

func CurrentClaims(c *fiber.Ctx) (sharedjwt.AccessTokenClaims, bool) {
	if c == nil {
		return sharedjwt.AccessTokenClaims{}, false
	}

	claims, ok := c.Locals(claimsLocalKey).(sharedjwt.AccessTokenClaims) //верно понял что c.Locals берет значение по ключу? если так, то в какой момент мы вообще задали эту мапу? и далее, мы тут делаем type assertion? и отдельно напомни когда надо с nil делать type assertion и как?
	return claims, ok
}

func CurrentPrincipal(c *fiber.Ctx) (sharedauth.Principal, bool) { //в чем разница между контекстом и принципалом? почему принципал не используется у меня?
	if c == nil {
		return sharedauth.Principal{}, false
	}

	return sharedauth.PrincipalFromContext(c.UserContext())
}

func bearerToken(header string) (string, bool) {
	const prefix = "bearer "

	header = strings.TrimSpace(header)
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}

	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}

func contextWithClaims(parent context.Context, claims sharedjwt.AccessTokenClaims) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	ctx := context.WithValue(parent, userIDContextKey, claims.UserID)
	ctx = context.WithValue(ctx, deviceIDContextKey, claims.DeviceID)
	ctx = context.WithValue(ctx, roleContextKey, claims.Role)

	return ctx
}
