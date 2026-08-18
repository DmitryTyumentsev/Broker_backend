// [БОЙЛЕРПЛЕЙТ] Отдельный бинарник для миграций схемы app.
// Запуск: go run ./services/app/authservice/cmd/migrator
//
// Почему отдельно от сервиса: миграции накатываются один раз при деплое
// (init-контейнер / CI-шаг), а не при каждом старте пода. Иначе три
// реплики начнут мигрировать одновременно.
//
// В настоящем гибриде этот каталог катил бы мигратор Laravel — здесь он
// стоит заглушкой, чтобы владелец схемы app был один и тот же.

package main

import (
	"Broker_backend/services/app/authservice/internal/config"
	"context"
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
		migrationsDir = filepath.Join("migrations", "app")
	}

	// goose работает через database/sql, а не через pgx-пул напрямую:
	// ему нужны обычные Exec/Query без пулинга.
	db, err := sql.Open("pgx", pg.ConnectionString())
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = db.Close() }()

	pingCtx, cancelPing := context.WithTimeout(context.Background(), pg.ConnectTimeout)
	defer cancelPing()

	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	// Таблица версий своя у каждого владельца схемы: app.goose_db_version
	// и integration.goose_db_version. Одна на двоих означала бы, что второй
	// мигратор считает чужие версии уже накаченными.
	if tableName := pg.GooseTableName(); tableName != "" {
		goose.SetTableName(tableName)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
