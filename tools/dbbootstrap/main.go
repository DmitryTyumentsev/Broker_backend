// dbbootstrap — мигратор каталога migrations/bootstrap.
//
// Отдельный бинарь, потому что у него отдельные права: схемы, роли и
// гранты создаёт суперпользователь, а миграторы контуров ходят под
// app_user и go_user, у которых таких прав нет и быть не должно.
//
// Запуск:
//
//	BOOTSTRAP_DSN=postgres://postgres:postgres@localhost:5432/broker?sslmode=disable \
//	  go run ./tools/dbbootstrap
//
// Своя таблица версий — public.goose_bootstrap_version. Общая с контурами
// означала бы, что мигратор app считает версию 00001 уже накаченной.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib" // регистрирует драйвер "pgx" в database/sql
)

const (
	defaultDSN       = "postgres://postgres:postgres@localhost:5432/broker?sslmode=disable"
	versionTableName = "public.goose_bootstrap_version"
	pingTimeout      = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "dbbootstrap: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("bootstrap migrations applied successfully")
}

func run() error {
	dsn := strings.TrimSpace(os.Getenv("BOOTSTRAP_DSN"))
	if dsn == "" {
		dsn = defaultDSN
	}

	migrationsDir := strings.TrimSpace(os.Getenv("BOOTSTRAP_MIGRATIONS_PATH"))
	if migrationsDir == "" {
		migrationsDir = filepath.Join("migrations", "bootstrap")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = db.Close() }()

	// PingContext, а не Ping: без потолка ожидания шаг в CI зависает
	// вместо того, чтобы упасть.
	pingCtx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	goose.SetTableName(versionTableName)

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("run bootstrap migrations: %w", err)
	}

	return nil
}
