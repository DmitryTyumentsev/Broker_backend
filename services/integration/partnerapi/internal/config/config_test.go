package config

import (
	"strings"
	"testing"
	"time"
)

// Тест на модель authz: роли объявляются один раз в business.authz.roles,
// пермишены ссылаются только на объявленные роли. Опечатка в yaml должна
// падать на старте сервиса, а не на первом запросе в проде.
func TestConfigValidateRejectsPermissionRoleOutsideDeclaredRoles(t *testing.T) {
	cfg := validTestConfig()
	cfg.Business.Authz.Permissions["developer.admin.access"] = []string{"unexpected_admin"}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "unexpected_admin") {
		t.Fatalf("expected invalid role in error, got %v", err)
	}
}

func TestConfigValidateRejectsUnknownRole(t *testing.T) {
	cfg := validTestConfig()
	cfg.Business.Authz.Roles = append(cfg.Business.Authz.Roles, "chief_happiness_officer")

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "chief_happiness_officer") {
		t.Fatalf("expected unknown role in error, got %v", err)
	}
}

// Redis нужен только под лимиты и идемпотентность. Если и то и другое
// выключено — сервис обязан подниматься без Redis вообще.
func TestConfigValidateDoesNotRequireRedisWhenRedisMiddlewareDisabled(t *testing.T) {
	cfg := validTestConfig()
	cfg.Business.AuthRateLimit.Enabled = false
	cfg.Business.DefaultRateLimit.Enabled = false
	cfg.Business.Idempotency.Enabled = false
	cfg.Database.Redis = RedisConfig{}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate config: %v", err)
	}
}

func TestConfigValidateRequiresFixationGRPCAddress(t *testing.T) {
	cfg := validTestConfig()
	cfg.FixationGRPC.Address = ""

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for empty fixation_grpc.address")
	}
}

func validTestConfig() *Config {
	return &Config{
		Environment: "local",
		Server: ServerConfig{
			Port:           8080,
			BodyLimitBytes: 2 * 1024 * 1024,
		},
		Business: BusinessConfig{
			ContextTimeout:    time.Second,
			RequestTimeout:    8 * time.Second,
			OperationTimeout:  5 * time.Second,
			AccessTokenSecret: "test-access-secret-change-me-32-bytes-minimum",
			AccessTokenIssuer: "authservice",
			AuthRateLimit: RateLimitConfig{
				Enabled: true,
				Limit:   10,
				Window:  time.Minute,
			},
			DefaultRateLimit: RateLimitConfig{
				Enabled: true,
				Limit:   100,
				Window:  time.Minute,
			},
			Idempotency: IdempotencyConfig{
				Enabled:          true,
				TTL:              24 * time.Hour,
				Header:           "Idempotency-Key",
				LockPrefix:       "idem:api",
				MaxResponseBytes: 262144,
			},
			Authz: AuthzConfig{
				Roles: []string{"superadmin", "broker_team_member"},
				Permissions: map[string][]string{
					"api.protected.access":   {"superadmin", "broker_team_member"},
					"developer.admin.access": {"superadmin"},
				},
			},
		},
		AuthGRPC: AuthGRPCConfig{
			Address: "localhost:50051",
		},
		FixationGRPC: FixationGRPCConfig{
			Address: "localhost:50052",
		},
		Database: DatabaseConfig{
			Redis: RedisConfig{
				Host:         "127.0.0.1",
				Port:         6379,
				DialTimeout:  time.Second,
				ReadTimeout:  time.Second,
				WriteTimeout: time.Second,
			},
		},
	}
}
