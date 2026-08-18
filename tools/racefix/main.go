// racefix — генератор гонки на горячем пути.
//
// N горутин ОДНОВРЕМЕННО пытаются зафиксировать один и тот же телефон
// в одном и том же проекте. Печатает распределение ответов по кодам.
//
//	go run ./tools/racefix -n 50 -token "$TOKEN"
//
// Что здесь считается успехом. Ровно один 201 и N-1 конфликтов — значит
// уникальность держит база. Два 201 — значит два агентства получат
// комиссию за одного клиента, и это уже не баг в коде, а деньги.
// Пятисотки в выдаче — отдельный разговор: конфликт, доехавший до клиента
// как 500, неотличим для CRM агентства от аварии, и она будет ретраить.
//
// Барьер на старте обязателен. Без него первые горутины успевают
// отработать раньше, чем стартуют последние, гонки не случается,
// и скрипт зеленеет даже на сломанном индексе.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type request struct {
	FixFor    string `json:"fix_for"`
	Phone     string `json:"phone"`
	ProjectID string `json:"project_id"`
}

type outcome struct {
	status    int
	body      string
	transport error
	duration  time.Duration
}

// main только запускает и решает, что делать с ошибкой. Всё остальное
// в run(): os.Exit посреди функции обрывает defer, и отмена контекста
// не срабатывает.
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "racefix: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		baseURL    = flag.String("url", "http://localhost:8080", "адрес partnerapi")
		token      = flag.String("token", "", "access token брокера (обязателен)")
		goroutines = flag.Int("n", 50, "сколько горутин стартует одновременно")
		phone      = flag.String("phone", "+7 (999) 111-22-33", "телефон клиента")
		projectID  = flag.String("project", "", "uuid проекта (обязателен)")
		fixFor     = flag.String("fix-for", "", "uuid брокера, на кого фиксируем (обязателен)")
		idemKey    = flag.String("idempotency-key", "", "если задан — уходит один и тот же ключ у всех горутин")
		timeout    = flag.Duration("timeout", 30*time.Second, "потолок на весь прогон")
	)
	flag.Parse()

	if strings.TrimSpace(*token) == "" || strings.TrimSpace(*projectID) == "" || strings.TrimSpace(*fixFor) == "" {
		flag.Usage()

		return errors.New("нужны -token, -project и -fix-for; идентификаторы печатает `make seed`")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	payload, err := json.Marshal(request{
		FixFor:    *fixFor,
		Phone:     *phone,
		ProjectID: *projectID,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	endpoint := strings.TrimRight(*baseURL, "/") + "/api/v1/fixations"

	// Один клиент на всех: пул соединений общий, и это ближе к реальности,
	// чем N отдельных клиентов с одним коннектом каждый.
	// MaxConnsPerHost поднят — иначе горутины выстроятся в очередь
	// на транспорте, и никакой гонки на сервере не будет.
	client := &http.Client{
		Timeout: *timeout,
		Transport: &http.Transport{
			MaxIdleConns:        *goroutines * 2,
			MaxIdleConnsPerHost: *goroutines * 2,
			MaxConnsPerHost:     *goroutines * 2,
		},
	}

	// start — стартовый барьер. Все горутины блокируются на чтении из
	// незакрытого канала. close(start) отпускает их ОДНОВРЕМЕННО:
	// чтение из закрытого канала возвращается сразу и у всех.
	start := make(chan struct{})

	// Буфер на всех — иначе горутины залипнут на отправке, пока основная
	// не начнёт читать, и Wait никогда не вернётся.
	results := make(chan outcome, *goroutines)

	var wg sync.WaitGroup
	for i := 0; i < *goroutines; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			<-start

			startedAt := time.Now()
			status, body, err := doRequest(ctx, client, endpoint, *token, *idemKey, payload)
			results <- outcome{
				status:    status,
				body:      body,
				transport: err,
				duration:  time.Since(startedAt),
			}
		}()
	}

	fmt.Printf("стартуем %d горутин на %s\n", *goroutines, endpoint)
	fmt.Printf("телефон %q, проект %s\n\n", *phone, *projectID)

	wallStart := time.Now()
	close(start)
	wg.Wait()
	close(results)
	wall := time.Since(wallStart)

	report(results, *goroutines, wall)

	return nil
}

func doRequest(
	ctx context.Context,
	client *http.Client,
	endpoint, token, idemKey string,
	payload []byte,
) (int, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	if idemKey != "" {
		// Один ключ на всех — проверка идемпотентности, а не уникальности:
		// тогда ожидается один 201 и N-1 ответов «уже обрабатывается».
		req.Header.Set("Idempotency-Key", idemKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer func() { _ = resp.Body.Close() }()

	// Тело читаем целиком даже когда оно не нужно: иначе соединение
	// не вернётся в пул и следующая горутина откроет новое.
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	return resp.StatusCode, strings.TrimSpace(string(body)), nil
}

func report(results <-chan outcome, total int, wall time.Duration) {
	byStatus := make(map[int]int)
	samples := make(map[int]string)
	transportErrors := make(map[string]int)

	var slowest time.Duration

	for res := range results {
		if res.duration > slowest {
			slowest = res.duration
		}

		if res.transport != nil {
			transportErrors[res.transport.Error()]++
			continue
		}

		byStatus[res.status]++
		if _, seen := samples[res.status]; !seen {
			samples[res.status] = truncate(res.body, 160)
		}
	}

	codes := make([]int, 0, len(byStatus))
	for code := range byStatus {
		codes = append(codes, code)
	}
	sort.Ints(codes)

	fmt.Println("распределение ответов:")
	for _, code := range codes {
		fmt.Printf("  %d  %4d  %s\n", code, byStatus[code], statusHint(code))
		fmt.Printf("        пример: %s\n", samples[code])
	}

	if len(transportErrors) > 0 {
		fmt.Println("\nошибки транспорта (ответа не было вообще):")
		for message, count := range transportErrors {
			fmt.Printf("  %4d  %s\n", count, truncate(message, 160))
		}
	}

	fmt.Printf("\nвсего %d, весь прогон %s, самый долгий ответ %s\n", total, wall.Round(time.Millisecond), slowest.Round(time.Millisecond))

	switch created := byStatus[http.StatusCreated]; created {
	case 1:
		fmt.Println("\nсоздана ровно одна фиксация — уникальность держится")
	case 0:
		fmt.Println("\nне создано ни одной. Смотри тело ответа: скорее всего запрос не проходит проверки")
	default:
		fmt.Printf("\nсоздано %d фиксаций на один телефон в одном проекте.\n", created)
		fmt.Println("это значит, что комиссию за одного клиента получат несколько агентств")
	}
}

func statusHint(code int) string {
	switch code {
	case http.StatusCreated:
		return "фиксация создана"
	case http.StatusConflict:
		return "конфликт — телефон уже зафиксирован"
	case http.StatusUnauthorized:
		return "токен не принят"
	case http.StatusForbidden:
		return "роли не хватает прав"
	case http.StatusTooManyRequests:
		return "упёрлись в лимит partnerapi, а не в гонку"
	case http.StatusBadRequest:
		return "запрос не прошёл валидацию"
	case http.StatusInternalServerError:
		return "внутренняя ошибка — конфликт не должен выглядеть так"
	default:
		return ""
	}
}

func truncate(s string, limit int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= limit {
		return s
	}

	return s[:limit] + "…"
}
