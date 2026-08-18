// dbcheck — «куда я вообще подключился и что там лежит».
//
//	go run ./tools/dbcheck -dsn "postgres://postgres:postgres@127.0.0.1:55432/broker?sslmode=disable"
//
// Зачем отдельный инструмент. Мигратор ходит в базу с хоста
// (127.0.0.1:55432), а сидер — из сети compose (postgres:5432). Пока это
// один и тот же сервер, всё хорошо. Когда это разные серверы, картина
// выглядит абсурдно: `goose: no migrations to run, current version: 6`,
// и тут же `relation "integration.fixations" does not exist`.
//
// Диагноз в такой ситуации ставится сравнением: запускаем одно и то же
// с двух сторон и смотрим на inet_server_addr, размер базы и список
// таблиц. Если они разошлись — серверов два.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "dbcheck: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		dsn     = flag.String("dsn", defaultDSN(), "строка подключения")
		label   = flag.String("label", "", "подпись в выводе: откуда смотрим")
		require = flag.String("require", "", "таблицы через запятую, без которых считать состояние сломанным")
		quiet   = flag.Bool("quiet", false, "молчать, если всё требуемое на месте")
		timeout = flag.Duration("timeout", 10*time.Second, "потолок на запросы")
	)
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	db, err := sql.Open("pgx", *dsn)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	// Режим постусловия: после миграций проверяем, что объекты правда
	// появились. Запись в таблице версий goose — это утверждение, а не
	// доказательство: она относится к тому серверу, к которому подключился
	// мигратор, и молчит о том, тот ли это сервер.
	if strings.TrimSpace(*require) != "" {
		return requireTables(ctx, db, *dsn, strings.Split(*require, ","), *quiet)
	}

	if *label != "" {
		fmt.Printf("── %s ──\n", *label)
	}

	// Отпечаток сервера. system_identifier уникален для кластера: два
	// разных постгреса дадут разные значения, и это самый прямой ответ
	// на вопрос «это одна база или две».
	if err := printRow(ctx, db,
		`select current_database(), current_user, inet_server_addr()::text,
		        current_setting('server_version'),
		        (select system_identifier::text from pg_control_system())`,
		[]string{"база", "роль", "адрес сервера", "версия", "system_identifier"}); err != nil {
		return err
	}

	fmt.Println()

	if err := printList(ctx, db,
		`select nspname from pg_namespace
		  where nspname not like 'pg\_%' and nspname <> 'information_schema'
		  order by nspname`,
		"схемы"); err != nil {
		return err
	}

	if err := printList(ctx, db,
		`select schemaname || '.' || tablename from pg_tables
		  where schemaname in ('app', 'integration', 'public')
		  order by 1`,
		"таблицы"); err != nil {
		return err
	}

	// Версии goose из всех трёх таблиц. Если таблица есть, а объектов
	// из миграции нет — значит смотрим не на тот сервер.
	for _, table := range []string{
		"public.goose_bootstrap_version",
		"app.goose_db_version",
		"integration.goose_db_version",
	} {
		printGooseVersion(ctx, db, table)
	}

	return nil
}

// requireTables проверяет наличие каждой таблицы из списка и объясняет,
// что делать, если её нет.
func requireTables(ctx context.Context, db *sql.DB, dsn string, tables []string, quiet bool) error {
	missing := make([]string, 0, len(tables))

	for _, qualified := range tables {
		qualified = strings.TrimSpace(qualified)
		if qualified == "" {
			continue
		}

		schema, table, ok := strings.Cut(qualified, ".")
		if !ok {
			return fmt.Errorf("ожидал схему и таблицу через точку, получил %q", qualified)
		}

		var exists bool

		err := db.QueryRowContext(ctx,
			`select exists (
			     select 1 from pg_tables where schemaname = $1 and tablename = $2
			 )`, schema, table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("проверка %s: %w", qualified, err)
		}

		if !exists {
			missing = append(missing, qualified)
		}
	}

	if len(missing) == 0 {
		if !quiet {
			fmt.Printf("состояние базы в порядке: %s\n", strings.Join(tables, ", "))
		}

		return nil
	}

	return fmt.Errorf(
		"миграции отчитались об успехе, но таблиц нет: %s\n"+
			"  подключение: %s\n"+
			"  Это значит, что мигратор и потребитель смотрят на РАЗНЫЕ серверы\n"+
			"  либо миграции накатились в другую базу. Разбор: make db-check",
		strings.Join(missing, ", "), maskPassword(dsn))
}

// maskPassword прячет пароль в DSN: строку подключения печатают в лог,
// и утечка пароля туда — отдельная неприятность.
func maskPassword(dsn string) string {
	at := strings.LastIndex(dsn, "@")
	slashes := strings.Index(dsn, "//")

	if at < 0 || slashes < 0 || at < slashes {
		return dsn
	}

	credentials := dsn[slashes+2 : at]

	user, _, ok := strings.Cut(credentials, ":")
	if !ok {
		return dsn
	}

	return dsn[:slashes+2] + user + ":***" + dsn[at:]
}

func defaultDSN() string {
	if dsn := strings.TrimSpace(os.Getenv("DBCHECK_DSN")); dsn != "" {
		return dsn
	}

	return "postgres://postgres:postgres@127.0.0.1:55432/broker?sslmode=disable"
}

func printRow(ctx context.Context, db *sql.DB, query string, titles []string) error {
	values := make([]sql.NullString, len(titles))
	dest := make([]any, len(titles))

	for i := range values {
		dest[i] = &values[i]
	}

	if err := db.QueryRowContext(ctx, query).Scan(dest...); err != nil {
		return fmt.Errorf("описание сервера: %w", err)
	}

	for i, title := range titles {
		fmt.Printf("  %-18s %s\n", title, values[i].String)
	}

	return nil
}

func printList(ctx context.Context, db *sql.DB, query, title string) error {
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("%s: %w", title, err)
	}
	defer func() { _ = rows.Close() }()

	items := make([]string, 0, 16)

	for rows.Next() {
		var item string
		if err := rows.Scan(&item); err != nil {
			return fmt.Errorf("%s: %w", title, err)
		}

		items = append(items, item)
	}

	// rows.Err() обязателен: оборванная выборка иначе выглядит как пустая,
	// а «таблиц нет» — это ровно тот вывод, который мы тут проверяем.
	if err := rows.Err(); err != nil {
		return fmt.Errorf("%s: %w", title, err)
	}

	if len(items) == 0 {
		fmt.Printf("  %s: пусто\n", title)

		return nil
	}

	fmt.Printf("  %s (%d):\n", title, len(items))

	for _, item := range items {
		fmt.Printf("    %s\n", item)
	}

	return nil
}

func printGooseVersion(ctx context.Context, db *sql.DB, table string) {
	var version sql.NullInt64

	err := db.QueryRowContext(ctx, `select max(version_id) from `+table).Scan(&version)
	if err != nil {
		fmt.Printf("  %-30s нет таблицы версий\n", table)

		return
	}

	fmt.Printf("  %-30s версия %d\n", table, version.Int64)
}
