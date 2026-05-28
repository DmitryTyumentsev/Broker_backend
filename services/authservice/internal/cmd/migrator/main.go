package main

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"Broker_backend/services/authservice/internal/config"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	pg := cfg.Database.Postgres

	dsn := pg.ConnectionString()

	migrationsDir := strings.TrimSpace(pg.MigrationsPath)
	if migrationsDir == "" {
		migrationsDir = filepath.Join("services", "authservice", "migrations")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(fmt.Errorf("open db: %w", err))
	}
	defer func() {
		_ = db.Close()
	}()

	if err := db.Ping(); err != nil {
		panic(fmt.Errorf("ping db: %w", err))
	}

	if err := goose.SetDialect("postgres"); err != nil {
		panic(fmt.Errorf("set goose dialect: %w", err))
	}

	if tableName := pg.GooseTableName(); tableName != "" {
		goose.SetTableName(tableName)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		panic(fmt.Errorf("run migrations: %w", err))
	}

	fmt.Println("migrations applied successfully")
}
