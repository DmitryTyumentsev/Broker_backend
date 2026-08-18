package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// keyDelimiter — разделитель уровней вложенности внутри viper.
//
// По умолчанию это точка, и на нашем конфиге это ломается. У нас есть
// ключи, в имени которых точка — часть имени, а не вложенность:
//
//	business.authz.permissions:
//	  api.protected.access: [...]
//	  fixation.new:         [...]
//
// С разделителем-точкой viper разбирает "api.protected.access" на три
// уровня и получает map[string]any вместо списка ролей. Unmarshal падает
// на «expected type 'string', got unconvertible type 'map[string]interface {}'».
//
// Переименовать пермишены нельзя: их имена — это константы в
// shared/pkg/authz/permissions, и точка в них осмысленна. Поэтому меняем
// разделитель на последовательность, которой в наших ключах нет.
//
// Следствие: все ключи, которые мы передаём в SetDefault и Get, тоже
// пишутся через "::". В yaml ничего не меняется — там вложенность задаёт
// сам формат.
const keyDelimiter = "::"

func Load(serviceName string, out any) error {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "local"
	}

	v := viper.NewWithOptions(viper.KeyDelimiter(keyDelimiter))

	v.SetConfigName(env)
	v.SetConfigType("yaml")
	v.SetEnvPrefix(strings.ToUpper(strings.NewReplacer("-", "_").Replace(serviceName)))

	v.AddConfigPath(filepath.Join("services", serviceName, "configs"))
	v.AddConfigPath(filepath.Join("services", "app", serviceName, "configs"))
	v.AddConfigPath(filepath.Join("services", "integration", serviceName, "configs"))
	v.AddConfigPath("configs")
	v.AddConfigPath(".")

	// Имя переменной окружения собирается из пути ключа: разделитель
	// уровней меняется на подчёркивание. database::postgres::dsn →
	// PARTNERAPI_DATABASE_POSTGRES_DSN. Точку тоже заменяем — иначе
	// пермишен с точкой в имени дал бы переменную, которую нельзя набрать.
	v.SetEnvKeyReplacer(strings.NewReplacer(keyDelimiter, "_", ".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	_ = godotenv.Load(".env")
	_ = godotenv.Load(filepath.Join("services", serviceName, "configs", env+".env"))
	_ = godotenv.Load(filepath.Join("services", "app", serviceName, "configs", env+".env"))
	_ = godotenv.Load(filepath.Join("services", "integration", serviceName, "configs", env+".env"))

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read config file for service=%s env=%s: %w", serviceName, env, err)
	}

	if err := v.Unmarshal(out); err != nil {
		return fmt.Errorf("unmarshal config for service=%s env=%s: %w", serviceName, env, err)
	}

	return nil
}

// setDefaults — значения на случай, если ключа нет ни в yaml, ни в env.
// Ключи пишутся через keyDelimiter, а не через точку: см. комментарий выше.
func setDefaults(v *viper.Viper) {
	v.SetDefault("server::host", "0.0.0.0")
	v.SetDefault("server::port", 8080)
	v.SetDefault("server::read_timeout", "5s")
	v.SetDefault("server::write_timeout", "5s")
	v.SetDefault("server::idle_timeout", "60s")

	v.SetDefault("business::context_timeout", "5s")
	v.SetDefault("business::lifetime_access_token", "15m")
	v.SetDefault("business::lifetime_refresh_token", "720h")

	v.SetDefault("database::postgres::ssl_mode", "disable")
	v.SetDefault("database::postgres::max_connections", 10)
	v.SetDefault("database::postgres::read_timeout", "3s")
	v.SetDefault("database::postgres::write_timeout", "3s")
}
