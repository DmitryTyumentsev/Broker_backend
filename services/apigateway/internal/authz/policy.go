package authz

import (
	"Broker_backend/services/apigateway/internal/config"
	"errors"
	"fmt"
	"strings"
)

type RolePermissionPolicy struct {
	permissions map[string]map[string]struct{}
}

func NewRolePermissionPolicy(cfg config.AuthzConfig) (*RolePermissionPolicy, error) {
	if len(cfg.Permissions) == 0 {
		return nil, errors.New("permissions must not be empty")
	}

	policy := &RolePermissionPolicy{
		permissions: make(map[string]map[string]struct{}, len(cfg.Permissions)),
	}

	for permission, allowedRoles := range cfg.Permissions {
		permission = strings.TrimSpace(permission)
		if permission == "" {
			return nil, errors.New("permission name is empty")
		}

		roleSet := make(map[string]struct{}, len(allowedRoles))
		for _, role := range allowedRoles {
			role = strings.TrimSpace(role)
			if role == "" {
				continue
			}

			roleSet[role] = struct{}{}
		}

		if len(roleSet) == 0 {
			return nil, fmt.Errorf("permission %q has no roles", permission)
		}

		policy.permissions[permission] = roleSet
	}

	return policy, nil
}

func (p *RolePermissionPolicy) IsAllowed(role string, permission string) bool {
	if p == nil {
		return false
	}

	role = strings.TrimSpace(role)
	permission = strings.TrimSpace(permission)

	if role == "" || permission == "" {
		return false
	}

	roles, ok := p.permissions[permission]
	if !ok {
		return false
	}

	_, ok = roles[role]
	return ok
}
