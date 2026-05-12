package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if len(os.Args) < 3 {
		panic("usage: migrator <postgres_dsn> <migrations_dir>")
	}

	dsn := os.Args[1]
	migrationsDir := os.Args[2]

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(fmt.Errorf("open db: %w", err))
	}
	defer func() {
		_ = db.Close()
	}()

	if err := goose.SetDialect("postgres"); err != nil {
		panic(fmt.Errorf("set goose dialect: %w", err))
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		panic(fmt.Errorf("run migrations: %w", err))
	}
}
