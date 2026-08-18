package config

import (
	"testing"

	sharedconfig "Broker_backend/shared/pkg/config"
)

// Тест на настоящий configs/local.yaml, а не на структуру в памяти:
// разобранный руками конфиг не доказывает, что файл вообще читается.
func TestRealConfigFilesLoad(t *testing.T) {
	sharedconfig.ChdirToRepoRoot(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("configs/local.yaml не загрузился: %v", err)
	}

	// Контур монолита ходит в базу под app_user и владеет схемой app.
	// Если schema_name уедет, goose положит таблицу версий не туда.
	if cfg.Database.Postgres.SchemaName != "app" {
		t.Errorf("schema_name = %q, ожидали app", cfg.Database.Postgres.SchemaName)
	}

	if got := cfg.Database.Postgres.GooseTableName(); got != "app.goose_db_version" {
		t.Errorf("таблица версий goose = %q, ожидали app.goose_db_version", got)
	}
}
