package config

import (
	"testing"

	sharedconfig "Broker_backend/shared/pkg/config"
)

// Тест на НАСТОЯЩИЕ файлы из репозитория, а не на структуру в памяти.
//
// Регрессия: конфиг разбирался на структуру в других тестах и выглядел
// исправным, но LoadConfig на реальном local.yaml падал. Ключи пермишенов
// содержат точки (api.protected.access), а viper по умолчанию считает
// точку разделителем уровней — и вместо списка ролей получал вложенную
// карту. Сервис не поднимался вообще, а поймать это можно было только
// запуском.
//
// Отсюда правило: конфиги в репозитории обязаны грузиться в тестах.
func TestRealConfigFilesLoad(t *testing.T) {
	sharedconfig.ChdirToRepoRoot(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("configs/local.yaml не загрузился: %v", err)
	}

	// Пермишены — то место, где всё ломалось. Проверяем, что имя ключа
	// осталось целым, а значение действительно список ролей.
	permissions := cfg.Business.Authz.Permissions
	if len(permissions) == 0 {
		t.Fatal("business.authz.permissions пуст")
	}

	for _, permission := range []string{"api.protected.access", "fixation.new"} {
		roles, ok := permissions[permission]
		if !ok {
			t.Errorf("пермишен %q не разобрался: ключ съеден разделителем", permission)

			continue
		}

		if len(roles) == 0 {
			t.Errorf("пермишен %q без ролей", permission)
		}
	}

	// Секция http раньше читалась по ключу "grpc" из-за опечатки
	// в mapstructure-теге, и CORS молча оставался выключенным.
	if !cfg.HTTP.CORS.Enabled {
		t.Error("http.cors.enabled не прочитался из yaml")
	}

	if cfg.AuthGRPC.Address == cfg.FixationGRPC.Address {
		t.Errorf("auth_grpc и fixation_grpc указывают на один адрес: %s", cfg.AuthGRPC.Address)
	}
}
