package middleware

import (
	"Broker_backend/services/apigateway/internal/config"
	"Broker_backend/services/apigateway/internal/transport/http/httperr"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthzPolicy interface {
	IsAllowed(role string, permission string) bool
}

type RolePermissionPolicy struct {
	permissions map[string]map[string]struct{}
}

func NewRolePermissionPolicy(cfg config.AuthzConfig) (*RolePermissionPolicy, error) {
	p := &RolePermissionPolicy{
		permissions: make(map[string]map[string]struct{}, len(cfg.Permissions)),
	}

	for permission, allowedRoles := range cfg.Permissions {
		roleSet := make(map[string]struct{}, len(allowedRoles))

		for _, role := range allowedRoles {
			role = strings.TrimSpace(role)
			if role == "" {
				continue
			}

			roleSet[role] = struct{}{}
		}

		p.permissions[permission] = roleSet
	}

	return p, nil
}

func (p *RolePermissionPolicy) IsAllowed(role string, permission string) bool {
	if p == nil {
		return false
	}

	roleSet, ok := p.permissions[permission]
	if !ok {
		return false
	}

	_, ok = roleSet[role]
	return ok
}

func RequirePermission(policy AuthzPolicy, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := CurrentClaims(c)
		if !ok {
			return httperr.WriteUnauthorized(c, "auth context is missing")
		}

		if !policy.IsAllowed(claims.Role, permission) {
			return httperr.WriteForbidden(c, "insufficient permissions")
		}

		return c.Next()
	}
}
