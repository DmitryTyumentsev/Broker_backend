// mock-profitbase — заглушка Profitbase (шахматка).
//
//	GET  /api/v4/property              список лотов, пагинация по курсору
//	POST /api/v4/property/{id}/hold    поставить бронь, умеет отдавать конфликт
//
// Ломается по команде — см. tools/mockctl.
//
//	MOCK_ADDR=:9102 go run ./tools/mock-profitbase
//
// Конфликт брони — главное, ради чего этот мок существует. Бронь лота
// в чужой системе может не пройти, потому что кто-то успел раньше;
// наша реакция на это и есть предмет проверки. Второй hold того же лота
// отдаёт 409 всегда, без всяких настроек: это не отказ, а нормальная
// работа предметной области.
package main

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"Broker_backend/tools/mockctl"
)

type property struct {
	ID        int64   `json:"id"`
	ProjectID string  `json:"project_id"`
	Building  string  `json:"house_name"`
	Number    string  `json:"number"`
	Floor     int     `json:"floor"`
	Rooms     int     `json:"rooms_amount"`
	AreaM2    float64 `json:"area"`
	Price     int64   `json:"price"`
	Status    string  `json:"status"`

	// Кем занят лот. Пусто = свободен. Хранится, чтобы повторный hold
	// тем же агентством не считался конфликтом: повтор своего же
	// запроса — это идемпотентность, а не гонка.
	HeldBy    string `json:"held_by,omitempty"`
	HeldUntil int64  `json:"held_until,omitempty"`
}

type store struct {
	mu    sync.Mutex
	items map[int64]*property
}

// newStore наполняет шахматку при старте: пустой мок пагинацию
// не проверяет никак.
func newStore() *store {
	s := &store{items: make(map[int64]*property)}

	projects := []string{"pb-project-1", "pb-project-2"}
	statuses := []string{"free", "free", "free", "booked", "sold"}

	var id int64 = 1
	for p, projectID := range projects {
		for i := 1; i <= 120; i++ {
			floor := (i-1)/4 + 1
			rooms := (i-1)%4 + 1

			s.items[id] = &property{
				ID:        id,
				ProjectID: projectID,
				Building:  fmt.Sprintf("Корпус %d", (i-1)/60+1),
				Number:    strconv.Itoa(100 + i),
				Floor:     floor,
				Rooms:     rooms,
				AreaM2:    28.5 + float64(rooms)*12.4,
				Price:     int64(5_200_000 + rooms*1_800_000 + floor*40_000),
				Status:    statuses[(i+p)%len(statuses)],
			}
			id++
		}
	}

	return s
}

// list — курсорная пагинация: «отдай limit штук с id больше cursor».
// Курсор по id, а не offset: при offset вставка новой строки посреди
// выборки сдвигает всё, и клиент пропускает записи, не узнав об этом.
func (s *store) list(projectID string, cursor int64, limit int) ([]property, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	all := make([]property, 0, len(s.items))
	for _, item := range s.items {
		if item.ID <= cursor {
			continue
		}
		if projectID != "" && item.ProjectID != projectID {
			continue
		}
		all = append(all, *item)
	}

	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })

	if len(all) > limit {
		all = all[:limit]
	}

	var next int64
	if len(all) == limit && limit > 0 {
		next = all[len(all)-1].ID
	}

	return all, next
}

type holdResult int

const (
	holdOK holdResult = iota
	holdNotFound
	holdConflict
	holdRepeat // тот же владелец повторил запрос
)

func (s *store) hold(id int64, holder string, ttl time.Duration) (holdResult, *property) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return holdNotFound, nil
	}

	now := time.Now()

	// Протухшая бронь конфликтом не считается. Иначе один брошенный
	// запрос блокировал бы лот навсегда.
	held := item.HeldBy != "" && item.HeldUntil > now.Unix()

	switch {
	case held && item.HeldBy == holder:
		return holdRepeat, item
	case held:
		return holdConflict, item
	case item.Status == "sold":
		// Проданный лот забронировать нельзя вообще, кем бы ты ни был.
		return holdConflict, item
	}

	item.HeldBy = holder
	item.HeldUntil = now.Add(ttl).Unix()
	item.Status = "booked"

	return holdOK, item
}

func main() {
	addr := strings.TrimSpace(os.Getenv("MOCK_ADDR"))
	if addr == "" {
		addr = ":9102"
	}

	data := newStore()
	control := mockctl.New()

	mux := http.NewServeMux()
	control.Mount(mux)

	// Один префикс на оба маршрута: ServeMux в стандартной библиотеке
	// 2021 года не умеет параметры пути, поэтому {id} разбираем руками.
	mux.HandleFunc("/api/v4/property", control.Handler(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			mockctl.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{
				"error": "method not allowed",
			})
			return
		}

		handleList(data, w, r)
	}))

	mux.HandleFunc("/api/v4/property/", control.Handler(func(w http.ResponseWriter, r *http.Request) {
		handleHold(data, w, r)
	}))

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Fprintf(os.Stderr, "mock-profitbase listening on %s\n", addr)

	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "mock-profitbase: %v\n", err)
		os.Exit(1)
	}
}

func handleList(data *store, w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	cursor, _ := strconv.ParseInt(strings.TrimSpace(query.Get("cursor")), 10, 64)

	limit, err := strconv.Atoi(strings.TrimSpace(query.Get("limit")))
	if err != nil || limit <= 0 || limit > 200 {
		limit = 50
	}

	items, next := data.list(strings.TrimSpace(query.Get("project_id")), cursor, limit)

	response := map[string]any{
		"data":  items,
		"count": len(items),
	}

	if next > 0 {
		response["next_cursor"] = next
	}

	mockctl.WriteJSON(w, http.StatusOK, response)
}

func handleHold(data *store, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		mockctl.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	// /api/v4/property/{id}/hold
	rest := strings.TrimPrefix(r.URL.Path, "/api/v4/property/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) != 2 || parts[1] != "hold" {
		mockctl.WriteJSON(w, http.StatusNotFound, map[string]string{
			"error": "route not found",
		})
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		mockctl.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "property id must be integer",
		})
		return
	}

	// Кто бронирует. В настоящем Profitbase это связано с токеном;
	// здесь достаточно заголовка — важно само наличие владельца брони.
	holder := strings.TrimSpace(r.Header.Get("X-Agency-Id"))
	if holder == "" {
		holder = "unknown"
	}

	result, item := data.hold(id, holder, 15*time.Minute)

	switch result {
	case holdNotFound:
		mockctl.WriteJSON(w, http.StatusNotFound, map[string]string{
			"error": "property not found",
		})

	case holdConflict:
		// 409, а не 422 и не 400: запрос корректный, изменилось состояние
		// на той стороне. Клиент должен различать «я прислал ерунду»
		// и «меня опередили» — реакция на них разная.
		mockctl.WriteJSON(w, http.StatusConflict, map[string]any{
			"error":      "property is already held",
			"held_by":    item.HeldBy,
			"held_until": item.HeldUntil,
			"status":     item.Status,
		})

	case holdRepeat:
		// Повтор своего же запроса — 200, а не 409. Иначе любой ретрай
		// после таймаута выглядел бы как проигранная гонка.
		mockctl.WriteJSON(w, http.StatusOK, map[string]any{
			"data":   item,
			"repeat": true,
		})

	case holdOK:
		mockctl.WriteJSON(w, http.StatusOK, map[string]any{"data": item})
	}
}
