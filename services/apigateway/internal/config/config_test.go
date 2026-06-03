package config

import (
	"strings"
	"testing"
	"time"
)

func TestConfigValidateRejectsAdminRoleOutsideProtectedRoles(t *testing.T) {
	cfg := validTestConfig()
	cfg.Business.AdminAllowedRoles = []string{"superadmin", "unexpected_admin"}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "unexpected_admin") {
		t.Fatalf("expected invalid role in error, got %v", err)
	}
}

func TestConfigValidateDoesNotRequireRedisWhenRateLimitDisabled(t *testing.T) {
	cfg := validTestConfig()
	cfg.Business.AuthRateLimit.Enabled = false
	cfg.Business.DefaultRateLimit.Enabled = false
	cfg.Database.Redis = RedisConfig{}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate config: %v", err)
	}
}

func validTestConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
		},
		Business: BusinessConfig{
			ContextTimeout:        time.Second,
			AccessTokenSecret:     "test-access-secret-change-me-32-bytes-minimum",
			AccessTokenIssuer:     "authservice",
			ProtectedAllowedRoles: []string{"superadmin", "broker_team_member"},
			AdminAllowedRoles:     []string{"superadmin"},
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
		},
		AuthGRPC: AuthGRPCConfig{
			Address: "localhost:50051",
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
