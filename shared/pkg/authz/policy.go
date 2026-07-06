package authz

import (
	"strings"

	"Broker_backend/services/apigateway/internal/config" //ты уверен что правильно тут тянуть? из апи гейтвей сюда. ide показывает ошибку
)

type RolePermissionPolicy struct {
	permissions map[string]map[string]struct{}
}

func NewRolePermissionPolicy(cfg config.AuthzConfig) (*RolePermissionPolicy, error) {
	p := &RolePermissionPolicy{
		permissions: make(map[string]map[string]struct{}, len(cfg.Permissions)),
	}

	for permission, allowedRoles := range cfg.Permissions {
		permission = strings.TrimSpace(permission)

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
