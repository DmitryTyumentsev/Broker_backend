package config

import (
	"testing"

	sharedconfig "Broker_backend/shared/pkg/config"
)

// Тест на настоящий configs/local.yaml, а не на структуру в памяти.
func TestRealConfigFilesLoad(t *testing.T) {
	sharedconfig.ChdirToRepoRoot(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("configs/local.yaml не загрузился: %v", err)
	}

	// Go-контур владеет схемой integration и ходит под go_user.
	if cfg.Database.Postgres.SchemaName != "integration" {
		t.Errorf("schema_name = %q, ожидали integration", cfg.Database.Postgres.SchemaName)
	}

	if got := cfg.Database.Postgres.GooseTableName(); got != "integration.goose_db_version" {
		t.Errorf("таблица версий goose = %q, ожидали integration.goose_db_version", got)
	}

	// Срок фиксации — бизнес-параметр, и ноль здесь означает фиксацию,
	// протухшую в момент создания.
	if cfg.Business.FixationDuration <= 0 {
		t.Error("business.fixation_duration должен быть положительным")
	}

	// Соль для хэша телефона: без неё хэш вырождается в обычный SHA,
	// а телефон перебирается за минуты.
	if cfg.Business.HashSecret == "" {
		t.Error("business.hash_secret пуст")
	}
}
