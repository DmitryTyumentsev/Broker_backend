package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Validator interface {
	Validate() error
}

func Load(serviceName string, cfg any) error {
	v := viper.New()

	setup(v, serviceName)
	setDefaults(v, serviceName)
	if err := readConfigFile(v, serviceName); err != nil {
		return err
	}

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	if validatable, ok := cfg.(Validator); ok {
		if err := validatable.Validate(); err != nil {
			return fmt.Errorf("validate config: %w", err)
		}
	}

	return nil
}

func setup(v *viper.Viper, serviceName string) {
	v.SetConfigType("yaml")

	// env-prefix: APIGATEWAY_..., AUTHSERVICE_...
	v.SetEnvPrefix(strings.ToUpper(serviceName))

	// server.port -> APIGATEWAY_SERVER_PORT
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.AutomaticEnv()
}

func readConfigFile(v *viper.Viper, serviceName string) error {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "local"
	}

	configName := env
	v.SetConfigName(configName)

	paths := []string{
		filepath.Join("services", serviceName, "configs"),
		filepath.Join(".", "services", serviceName, "configs"),
		filepath.Join("configs"),
	}

	for _, p := range paths {
		v.AddConfigPath(p)
	}

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read config file for service=%s env=%s: %w", serviceName, env, err)
	}

	return nil
}

func setDefaults(v *viper.Viper, serviceName string) {
	switch serviceName {
	case "apigateway":
		v.SetDefault("server.host", "0.0.0.0")
		v.SetDefault("server.port", 8080)
		v.SetDefault("server.read_timeout", "15s")
		v.SetDefault("server.write_timeout", "15s")
		v.SetDefault("server.idle_timeout", "60s")
		v.SetDefault("business.context_timeout", "5s")
		v.SetDefault("auth_grpc.address", "localhost:50051")

	case "authservice":
		v.SetDefault("server.host", "0.0.0.0")
		v.SetDefault("server.port", 50051)
		v.SetDefault("business.context_timeout", "5s")

		v.SetDefault("database.postgres.ssl_mode", "disable")
		v.SetDefault("database.postgres.max_connections", 10)
		v.SetDefault("database.postgres.min_connections", 1)
		v.SetDefault("database.postgres.max_idle_time", "30m")
		v.SetDefault("database.postgres.max_lifetime", "60m")
		v.SetDefault("database.postgres.health_check_period", "5m")
		v.SetDefault("database.postgres.write_timeout", "30s")
		v.SetDefault("database.postgres.read_timeout", "15s")
		v.SetDefault("database.postgres.connect_timeout", "5s")
	}
}
