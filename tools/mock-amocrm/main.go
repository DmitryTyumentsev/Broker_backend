// mock-amocrm — заглушка amoCRM.
//
// Воспроизводит ровно то, что нужно нашей интеграции:
//
//	POST /api/v4/leads   создать сделку
//	GET  /api/v4/leads   выбрать изменённые с момента X, страницами по курсору
//
// И, главное, умеет ломаться по команде — см. tools/mockctl.
//
//	MOCK_ADDR=:9101 go run ./tools/mock-amocrm
//
// Сделки хранятся в памяти. Перезапуск = чистое состояние, и это
// осознанно: тест, который зависит от данных прошлого прогона, врёт.
package main

import (
	"encoding/json"
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

type lead struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Price    int64  `json:"price"`
	StatusID int64  `json:"status_id"`
	Phone    string `json:"phone,omitempty"`
	AgencyID string `json:"agency_id,omitempty"`

	// FixationID — наш идентификатор фиксации, который интеграция кладёт
	// в сделку при создании. В настоящем amoCRM это было бы кастомное поле.
	// Без него сверить наши фиксации с их сделками нечем: телефон
	// в integration.fixations хранится только хэшем.
	FixationID string `json:"fixation_id,omitempty"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

type store struct {
	mu     sync.Mutex
	leads  map[int64]*lead
	nextID int64
}

func newStore() *store {
	return &store{
		leads:  make(map[int64]*lead),
		nextID: 1000,
	}
}

func (s *store) create(in lead) lead {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	now := time.Now().Unix()

	created := in
	created.ID = s.nextID
	created.CreatedAt = now
	created.UpdatedAt = now

	s.leads[created.ID] = &created

	return created
}

// list отдаёт сделки, изменённые не раньше updatedFrom, начиная с
// курсора afterID, не больше limit штук.
//
// Сортировка по (updated_at, id), а не только по updated_at: у сделок,
// изменённых в одну секунду, порядок иначе недетерминирован, и страницы
// начинают терять и дублировать записи. Это классический баг курсорной
// пагинации, и здесь он воспроизводится специально правильно — чтобы
// расхождения, которые ты будешь искать, были настоящими.
func (s *store) list(updatedFrom int64, afterID int64, limit int) ([]lead, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	all := make([]lead, 0, len(s.leads))
	for _, l := range s.leads {
		if l.UpdatedAt >= updatedFrom && l.ID > afterID {
			all = append(all, *l)
		}
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].UpdatedAt != all[j].UpdatedAt {
			return all[i].UpdatedAt < all[j].UpdatedAt
		}
		return all[i].ID < all[j].ID
	})

	if len(all) > limit {
		all = all[:limit]
	}

	var next int64
	if len(all) == limit && limit > 0 {
		next = all[len(all)-1].ID
	}

	return all, next
}

func main() {
	addr := strings.TrimSpace(os.Getenv("MOCK_ADDR"))
	if addr == "" {
		addr = ":9101"
	}

	data := newStore()
	control := mockctl.New()

	mux := http.NewServeMux()
	control.Mount(mux)

	mux.HandleFunc("/api/v4/leads", control.Handler(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreateLead(data, w, r)
		case http.MethodGet:
			handleListLeads(data, w, r)
		default:
			mockctl.WriteJSON(w, http.StatusMethodNotAllowed, map[string]string{
				"error": "method not allowed",
			})
		}
	}))

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Fprintf(os.Stderr, "mock-amocrm listening on %s\n", addr)

	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "mock-amocrm: %v\n", err)
		os.Exit(1)
	}
}

func handleCreateLead(data *store, w http.ResponseWriter, r *http.Request) {
	// amoCRM принимает МАССИВ сделок даже при создании одной штуки —
	// воспроизводим как есть. Клиент, написанный под одиночный объект,
	// должен здесь споткнуться, а не в проде.
	var batch []lead
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		mockctl.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "expected array of leads: " + err.Error(),
		})
		return
	}

	if len(batch) == 0 {
		mockctl.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "empty batch",
		})
		return
	}

	created := make([]lead, 0, len(batch))
	for _, in := range batch {
		created = append(created, data.create(in))
	}

	mockctl.WriteJSON(w, http.StatusOK, map[string]any{
		"_embedded": map[string]any{"leads": created},
	})
}

func handleListLeads(data *store, w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// Фильтр по времени изменения — то, на чём стоит вся инкрементальная
	// синхронизация. Формат ключа тот же, что у настоящего amoCRM.
	updatedFrom, err := parseInt(query.Get("filter[updated_at][from]"), 0)
	if err != nil {
		mockctl.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "filter[updated_at][from] must be unix timestamp",
		})
		return
	}

	cursor, err := parseInt(query.Get("cursor"), 0)
	if err != nil {
		mockctl.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "cursor must be integer",
		})
		return
	}

	limit, err := parseInt(query.Get("limit"), 50)
	if err != nil || limit <= 0 || limit > 250 {
		limit = 50
	}

	leads, next := data.list(updatedFrom, cursor, int(limit))

	// 204 на пустую выборку — тоже как у настоящего amoCRM.
	// Клиент, который ждёт JSON всегда, здесь получит пустое тело.
	if len(leads) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := map[string]any{
		"_embedded": map[string]any{"leads": leads},
	}

	// Курсор отдаём, только если страница полная. Пустой next — сигнал
	// «данные кончились»; выдавать его на каждой странице означало бы
	// заставить клиента ходить за пустотой вечно.
	if next > 0 {
		response["_links"] = map[string]any{
			"next": map[string]string{
				"href": fmt.Sprintf("/api/v4/leads?cursor=%d&limit=%d", next, limit),
			},
		}
		response["next_cursor"] = next
	}

	mockctl.WriteJSON(w, http.StatusOK, response)
}

func parseInt(raw string, fallback int64) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}

	return strconv.ParseInt(raw, 10, 64)
}
