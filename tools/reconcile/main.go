// reconcile — сверка наших фиксаций с тем, что лежит в amoCRM.
//
//	go run ./tools/reconcile -amocrm http://localhost:9101
//
// Зачем это нужно. Две системы синхронизируются через очередь и ретраи,
// то есть согласованы не в моменте, а «в конце концов». Вопрос «почему
// у агентства фиксация есть, а сделки в CRM нет» приходит регулярно,
// и отвечать на него по логам — значит не ответить.
//
// Скрипт даёт три списка:
//
//	только у нас   — фиксация есть, сделки нет. Обычно застрял outbox
//	                 или запросы в amoCRM падали и ретраи кончились
//	только в CRM   — сделка есть, фиксации нет. Либо её завели руками,
//	                 либо мы удалили свою, либо отправили дважды
//	расходится     — обе есть, но статусы разные. Самый неприятный случай:
//	                 обе стороны считают себя правыми
//
// Читает базу ТОЛЬКО на select и ничего не чинит: это инструмент
// расследования. Автоматическая «синхронизация одной кнопкой» в системе,
// где от статуса зависят деньги, — способ размножить ошибку.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type fixation struct {
	ID        string
	Status    string
	AgencyID  string
	ProjectID string
	FixedAt   time.Time
	ExpiresAt time.Time
}

type lead struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	StatusID   int64  `json:"status_id"`
	AgencyID   string `json:"agency_id"`
	FixationID string `json:"fixation_id"`
	UpdatedAt  int64  `json:"updated_at"`
}

type leadsPage struct {
	Embedded struct {
		Leads []lead `json:"leads"`
	} `json:"_embedded"`
	NextCursor int64 `json:"next_cursor"`
}

const maxPages = 10000

// main только запускает и решает, что делать с ошибкой. Всё остальное
// в run(), иначе os.Exit посреди функции обрывает defer — и контекст
// с пулом соединений остаются неубранными.
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "reconcile: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		dsn      = flag.String("dsn", defaultDSN(), "DSN к базе, роль go_user")
		amocrm   = flag.String("amocrm", "http://localhost:9101", "адрес amoCRM (мока)")
		since    = flag.Duration("since", 180*24*time.Hour, "насколько назад смотреть")
		pageSize = flag.Int("page-size", 100, "размер страницы при чтении сделок")
		verbose  = flag.Bool("v", false, "печатать все расхождения, а не первые 20")
		timeout  = flag.Duration("timeout", 60*time.Second, "потолок на весь прогон")
	)
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	ours, err := loadFixations(ctx, *dsn, *since)
	if err != nil {
		return fmt.Errorf("чтение фиксаций: %w", err)
	}

	theirs, err := loadLeads(ctx, *amocrm, time.Now().Add(-*since), *pageSize)
	if err != nil {
		return fmt.Errorf("чтение сделок amoCRM: %w", err)
	}

	report(ours, theirs, *verbose)

	return nil
}

func defaultDSN() string {
	if dsn := strings.TrimSpace(os.Getenv("RECONCILE_DSN")); dsn != "" {
		return dsn
	}

	return "postgres://go_user:go_user@localhost:55432/broker?sslmode=disable&search_path=integration,public"
}

func loadFixations(ctx context.Context, dsn string, since time.Duration) (map[string]fixation, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()

	// Схема указана явно, хотя в DSN есть search_path: запрос не должен
	// зависеть от того, каким соединением его выполнили.
	rows, err := pool.Query(ctx, `
		SELECT id::text, status, agency_id::text, project_id::text, fixed_at, expires_at
		  FROM integration.fixations
		 WHERE fixed_at >= $1
		 ORDER BY fixed_at`, time.Now().Add(-since))
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	out := make(map[string]fixation)

	for rows.Next() {
		var f fixation
		if err := rows.Scan(&f.ID, &f.Status, &f.AgencyID, &f.ProjectID, &f.FixedAt, &f.ExpiresAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		out[f.ID] = f
	}

	// rows.Err() обязателен. Без него выборка, оборванная на середине,
	// выглядит как «данных больше нет» — и отчёт врёт в самую опасную
	// сторону: показывает расхождения там, где их нет.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return out, nil
}

func loadLeads(ctx context.Context, baseURL string, since time.Time, pageSize int) (map[string]lead, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	out := make(map[string]lead)
	cursor := int64(0)

	for page := 0; ; page++ {
		// Потолок на число страниц: защита от API, которое отдаёт курсор
		// бесконечно. Без него скрипт молча уходит в вечный цикл.
		if page > maxPages {
			return nil, fmt.Errorf("слишком много страниц, курсор не сходится")
		}

		query := url.Values{}
		query.Set("filter[updated_at][from]", strconv.FormatInt(since.Unix(), 10))
		query.Set("limit", strconv.Itoa(pageSize))

		if cursor > 0 {
			query.Set("cursor", strconv.FormatInt(cursor, 10))
		}

		endpoint := strings.TrimRight(baseURL, "/") + "/api/v4/leads?" + query.Encode()

		next, err := fetchPage(ctx, client, endpoint, page, out)
		if err != nil {
			return nil, err
		}

		if next == 0 {
			break
		}

		cursor = next
	}

	return out, nil
}

func fetchPage(
	ctx context.Context,
	client *http.Client,
	endpoint string,
	page int,
	out map[string]lead,
) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("страница %d: %w", page, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// 204 — «изменений нет». Это конец выборки, а не ошибка.
	if resp.StatusCode == http.StatusNoContent {
		return 0, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))

		return 0, fmt.Errorf("страница %d: amoCRM ответил %d: %s",
			page, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var decoded leadsPage
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return 0, fmt.Errorf("страница %d: разбор ответа: %w", page, err)
	}

	for _, l := range decoded.Embedded.Leads {
		// Сделки без нашего идентификатора попадают в отчёт отдельной
		// строкой: связать их с фиксацией нечем, и это само по себе
		// находка, а не мусор в данных.
		key := l.FixationID
		if key == "" {
			key = fmt.Sprintf("lead-%d-without-fixation-id", l.ID)
		}

		out[key] = l
	}

	return decoded.NextCursor, nil
}

func report(ours map[string]fixation, theirs map[string]lead, verbose bool) {
	var onlyOurs, onlyTheirs, mismatched []string

	for id, f := range ours {
		l, ok := theirs[id]
		if !ok {
			onlyOurs = append(onlyOurs, fmt.Sprintf("%s  status=%s  agency=%s  fixed_at=%s",
				id, f.Status, short(f.AgencyID), f.FixedAt.Format(time.RFC3339)))

			continue
		}

		if expected := leadStatusFor(f.Status); expected != 0 && l.StatusID != expected {
			mismatched = append(mismatched, fmt.Sprintf(
				"%s  у нас=%s (ждали status_id=%d)  в amoCRM status_id=%d  lead=%d",
				id, f.Status, expected, l.StatusID, l.ID))
		}
	}

	for id, l := range theirs {
		if _, ok := ours[id]; !ok {
			onlyTheirs = append(onlyTheirs, fmt.Sprintf("lead=%d  fixation_id=%q  agency=%s",
				l.ID, l.FixationID, short(l.AgencyID)))
		}
	}

	sort.Strings(onlyOurs)
	sort.Strings(onlyTheirs)
	sort.Strings(mismatched)

	fmt.Printf("наших фиксаций: %d, сделок в amoCRM: %d\n\n", len(ours), len(theirs))

	printBlock("только у нас (в amoCRM не доехало)", onlyOurs, verbose)
	printBlock("только в amoCRM (нашей фиксации нет)", onlyTheirs, verbose)
	printBlock("расходится статус", mismatched, verbose)

	if len(onlyOurs) == 0 && len(onlyTheirs) == 0 && len(mismatched) == 0 {
		fmt.Println("расхождений нет")
	}
}

// leadStatusFor — соответствие наших статусов воронке amoCRM.
// Отдельная функция, а не константы по месту: соответствие меняется
// вместе с настройкой воронки у застройщика, и менять его надо в одном
// месте. 0 означает «нам всё равно, что там за статус».
func leadStatusFor(status string) int64 {
	switch status {
	case "active":
		return 142
	case "converted":
		return 143
	default:
		return 0
	}
}

func printBlock(title string, items []string, verbose bool) {
	fmt.Printf("%s: %d\n", title, len(items))

	limit := len(items)
	if !verbose && limit > 20 {
		limit = 20
	}

	for _, item := range items[:limit] {
		fmt.Printf("  %s\n", item)
	}

	if limit < len(items) {
		fmt.Printf("  ... ещё %d, запусти с -v\n", len(items)-limit)
	}

	fmt.Println()
}

func short(id string) string {
	if len(id) <= 8 {
		return id
	}

	return id[:8]
}
