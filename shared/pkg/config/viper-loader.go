package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func Load(serviceName string, out any) error {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "local"
	}

	v := viper.New()

	v.SetConfigName(env)
	v.SetConfigType("yaml")
	v.SetEnvPrefix(strings.ToUpper(strings.NewReplacer("-", "_").Replace(serviceName)))

	v.AddConfigPath(filepath.Join("services", serviceName, "configs"))
	v.AddConfigPath(filepath.Join("configs"))
	v.AddConfigPath(".")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	_ = godotenv.Load(".env")
	_ = godotenv.Load(filepath.Join("services", serviceName, "configs", env+".env"))

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read config file for service=%s env=%s: %w", serviceName, env, err)
	}

	if err := v.Unmarshal(out); err != nil {
		return fmt.Errorf("unmarshal config for service=%s env=%s: %w", serviceName, env, err)
	}

	return nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "5s")
	v.SetDefault("server.write_timeout", "5s")
	v.SetDefault("server.idle_timeout", "60s")

	v.SetDefault("business.context_timeout", "5s")
	v.SetDefault("business.lifetime_access_token", "15m")
	v.SetDefault("business.lifetime_refresh_token", "720h")

	v.SetDefault("database.postgres.ssl_mode", "disable")
	v.SetDefault("database.postgres.max_connections", 10)
	v.SetDefault("database.postgres.read_timeout", "3s")
	v.SetDefault("database.postgres.write_timeout", "3s")
}
