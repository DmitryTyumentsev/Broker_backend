// standenv — собирает окружение Postman из того, что реально лежит в базе.
//
//	go run ./tools/standenv            # пишет deploy/postman/broker.postman_environment.json
//
// Зачем. Идентификаторы агентств, проектов и брокеров генерирует сидер,
// и каждый прогон даёт новые. Переносить их в Postman руками — это шесть
// copy-paste и один шанс из шести ошибиться так, что потом полчаса
// объясняешь себе 404.
//
// Инструмент подбирает значения так, чтобы ими сразу можно было
// воспроизводить граничные случаи:
//
//	agencyId, brokerEmail, brokerId  — активное агентство и его сотрудник
//	foreignBrokerId                  — сотрудник ДРУГОГО агентства
//	projectId                        — активный проект
//	archivedProjectId                — архивный проект
//
// Выбор детерминированный (сортировка по имени), чтобы два запуска
// подряд давали одно и то же и диффы не шумели.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type variable struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

type environment struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	Values               []variable `json:"values"`
	PostmanVariableScope string     `json:"_postman_variable_scope"`
	ExportedAt           string     `json:"_postman_exported_at"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "standenv: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		dsn     = flag.String("dsn", defaultDSN(), "строка подключения к базе стенда")
		out     = flag.String("out", filepath.Join("deploy", "postman", "broker.postman_environment.json"), "кудаписать файл")
		timeout = flag.Duration("timeout", 15*time.Second, "потолок на запросы")
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

	values, err := collect(ctx, db)
	if err != nil {
		return err
	}

	env := environment{
		ID:                   "broker-stand",
		Name:                 "ЛК брокера — стенд",
		Values:               values,
		PostmanVariableScope: "environment",
		ExportedAt:           time.Now().UTC().Format(time.RFC3339),
	}

	raw, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	if err := os.WriteFile(*out, append(raw, '\n'), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", *out, err)
	}

	fmt.Printf("окружение записано: %s\n\n", *out)

	for _, v := range values {
		fmt.Printf("  %-20s %s\n", v.Key, v.Value)
	}

	fmt.Println("\nВ Postman: Import -> этот файл, потом выбрать окружение в списке справа сверху.")

	return nil
}

func collect(ctx context.Context, db *sql.DB) ([]variable, error) {
	// Базовые адреса и логин кладём сюда же: тогда окружение
	// самодостаточное и коллекция работает сразу после импорта.
	values := []variable{
		{Key: "partnerapi", Value: "http://localhost:8080"},
		{Key: "monolith", Value: "http://localhost:8000"},
		{Key: "amocrm", Value: "http://localhost:9101"},
		{Key: "profitbase", Value: "http://localhost:9102"},
		{Key: "password", Value: "password"},
		{Key: "deviceId", Value: "postman"},
		{Key: "phone", Value: "+7 (999) 111-22-33"},
	}

	// Агентство и «свой» брокер. Берём активное агентство: у заблокированного
	// свои сценарии, и делать его основным — значит на каждом запросе
	// упираться в чужую проверку.
	var agencyID, brokerID, brokerEmail string

	err := db.QueryRowContext(ctx, `
		select a.id::text, u.id::text, u.email
		  from app.agencies a
		  join app.users u on u.agency_id = a.id
		 where a.status = 'active'
		 order by a.name, u.email
		 limit 1`).Scan(&agencyID, &brokerID, &brokerEmail)
	if err != nil {
		return nil, describeEmpty(err, "активное агентство с сотрудником")
	}

	// Брокер из другого агентства — под сценарий «фиксирую на чужого».
	var foreignBrokerID string

	err = db.QueryRowContext(ctx, `
		select u.id::text
		  from app.users u
		 where u.agency_id is not null and u.agency_id <> $1::uuid
		 order by u.email
		 limit 1`, agencyID).Scan(&foreignBrokerID)
	if err != nil {
		return nil, describeEmpty(err, "сотрудник другого агентства")
	}

	projectID, err := oneProject(ctx, db, "active")
	if err != nil {
		return nil, err
	}

	archivedProjectID, err := oneProject(ctx, db, "archived")
	if err != nil {
		return nil, err
	}

	values = append(values,
		variable{Key: "agencyId", Value: agencyID},
		variable{Key: "brokerId", Value: brokerID},
		variable{Key: "brokerEmail", Value: brokerEmail},
		variable{Key: "foreignBrokerId", Value: foreignBrokerID},
		variable{Key: "projectId", Value: projectID},
		variable{Key: "archivedProjectId", Value: archivedProjectID},
	)

	for i := range values {
		values[i].Type = "default"
		values[i].Enabled = true
	}

	return values, nil
}

func oneProject(ctx context.Context, db *sql.DB, status string) (string, error) {
	var id string

	err := db.QueryRowContext(ctx,
		`select id::text from app.projects where status = $1 order by name limit 1`,
		status).Scan(&id)
	if err != nil {
		return "", describeEmpty(err, "проект со статусом "+status)
	}

	return id, nil
}

// describeEmpty превращает «нет строк» в понятную инструкцию:
// пустая база означает не поломку инструмента, а незапущенный сидер.
func describeEmpty(err error, what string) error {
	if strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
		return fmt.Errorf("в базе не найден %s — запусти `make seed`", what)
	}

	return fmt.Errorf("%s: %w", what, err)
}

func defaultDSN() string {
	if dsn := strings.TrimSpace(os.Getenv("STANDENV_DSN")); dsn != "" {
		return dsn
	}

	return "postgres://postgres:postgres@localhost:55432/broker?sslmode=disable"
}
