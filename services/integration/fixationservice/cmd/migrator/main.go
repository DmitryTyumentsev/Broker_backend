// [БОЙЛЕРПЛЕЙТ] Отдельный бинарник для миграций.
// Запуск: go run ./services/fixationservice/internal/cmd/migrator
//
// Почему отдельно от сервиса: миграции накатываются один раз при деплое
// (init-контейнер / CI-шаг), а не при каждом старте пода. Иначе три
// реплики начнут мигрировать одновременно.

package main

import (
	"Broker_backend/services/integration/fixationservice/internal/config"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib" // регистрирует драйвер "pgx" в database/sql
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "migrator: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("migrations applied successfully")
}

func run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	pg := cfg.Database.Postgres

	migrationsDir := strings.TrimSpace(pg.MigrationsPath)
	if migrationsDir == "" {
		migrationsDir = filepath.Join("migrations", "integration")
	}

	// goose работает через database/sql, а не через pgx-пул напрямую:
	// ему нужны обычные Exec/Query без пулинга.
	db, err := sql.Open("pgx", pg.ConnectionString())
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if tableName := pg.GooseTableName(); tableName != "" {
		goose.SetTableName(tableName)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
