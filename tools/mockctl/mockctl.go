// Package mockctl — общая часть моков внешних систем: управляемые отказы.
//
// Смысл всей затеи. Внешняя система в проде не «работает» и не «лежит».
// Она отвечает медленно, отдаёт 429 с Retry-After, роняет часть запросов
// в 500 и иногда исчезает на минуту целиком. Поведение интеграции в этих
// режимах — то, ради чего пишут ретраи, бэкофф и circuit breaker, и
// проверить его можно только на моке, который умеет их воспроизводить.
//
// Управление — через /_control у каждого мока:
//
//	# 300 мс задержки, 20% ответов 429 с Retry-After: 5, 10% пятисоток
//	curl -X POST localhost:9101/_control -d '{
//	  "delay_ms":300, "rate_429":0.2, "retry_after_s":5, "rate_500":0.1
//	}'
//
//	# упасть целиком на 30 секунд
//	curl -X POST localhost:9101/_control -d '{"down_for_s":30}'
//
//	# посмотреть текущее состояние и счётчики
//	curl localhost:9101/_control/state
//
//	# вернуть всё как было
//	curl -X POST localhost:9101/_control/reset
package mockctl

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Settings — что именно ломаем. Доли задаются как 0..1.
type Settings struct {
	// DelayMs — задержка перед ответом. Так проверяются клиентские
	// таймауты: без задержки любой таймаут выглядит «настроен правильно».
	DelayMs int `json:"delay_ms"`

	// Rate429 — доля ответов «слишком много запросов».
	// У amoCRM это штатный режим на пике, а не авария.
	Rate429     float64 `json:"rate_429"`
	RetryAfterS int     `json:"retry_after_s"`

	// Rate500 — доля ответов «у нас всё сломалось».
	// Принципиально отличается от 429: повторять можно, но не факт,
	// что запрос не выполнился на той стороне.
	Rate500 float64 `json:"rate_500"`

	// DownForS — при записи включает полное падение на N секунд.
	// В ответе /_control/state отдаётся остаток.
	DownForS int `json:"down_for_s"`
}

// Stats — счётчики, чтобы после прогона было чем подтвердить,
// что клиент действительно получил то, что мы ему подсунули.
type Stats struct {
	Total    int64 `json:"total"`
	OK       int64 `json:"ok"`
	TooMany  int64 `json:"too_many_requests"`
	Internal int64 `json:"internal_error"`
	Dropped  int64 `json:"dropped"`
}

// Controller — состояние мока. Один на процесс.
type Controller struct {
	mu        sync.Mutex
	settings  Settings
	downUntil time.Time
	stats     Stats
	rnd       *rand.Rand
}

func New() *Controller {
	return &Controller{
		// Свой источник случайности, а не глобальный rand: глобальный
		// шарится со всем процессом, и настройка «10% ошибок» в тесте
		// зависела бы от того, кто ещё его дёргал.
		rnd: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Handler оборачивает рабочий обработчик слоем управляемых отказов.
// Порядок проверок не случайный: сначала «нас нет вообще», потом
// задержка, потом коды ответа. Лежащая система не отвечает 429.
func (c *Controller) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.mu.Lock()
		settings := c.settings
		down := time.Now().Before(c.downUntil)
		roll429 := c.rnd.Float64()
		roll500 := c.rnd.Float64()
		c.stats.Total++
		c.mu.Unlock()

		if down {
			c.count(func(s *Stats) { s.Dropped++ })
			// Рвём соединение, а не отвечаем 503. Это разные аварии:
			// на 503 у клиента есть HTTP-ответ и код, а здесь он получит
			// EOF на середине — именно так выглядит упавший upstream,
			// и именно этот путь в клиенте чаще всего не обработан.
			hijackAndClose(w)
			return
		}

		if settings.DelayMs > 0 {
			time.Sleep(time.Duration(settings.DelayMs) * time.Millisecond)
		}

		if settings.Rate429 > 0 && roll429 < settings.Rate429 {
			c.count(func(s *Stats) { s.TooMany++ })

			retryAfter := settings.RetryAfterS
			if retryAfter <= 0 {
				retryAfter = 1
			}

			// Retry-After — не украшение. Клиент, который его игнорирует
			// и долбится сразу, получает следующий 429 и так по кругу.
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			writeJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "too many requests",
			})
			return
		}

		if settings.Rate500 > 0 && roll500 < settings.Rate500 {
			c.count(func(s *Stats) { s.Internal++ })
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "internal error",
			})
			return
		}

		c.count(func(s *Stats) { s.OK++ })
		next(w, r)
	}
}

// Mount вешает ручки управления. Их специально не оборачивают в Handler:
// управление обязано работать, даже когда мок «лежит», иначе поднять
// его обратно будет нечем.
func (c *Controller) Mount(mux *http.ServeMux) {
	mux.HandleFunc("/_control/state", func(w http.ResponseWriter, r *http.Request) {
		c.mu.Lock()
		settings := c.settings
		stats := c.stats
		downLeft := 0
		if remaining := time.Until(c.downUntil); remaining > 0 {
			downLeft = int(remaining.Seconds()) + 1
		}
		c.mu.Unlock()

		settings.DownForS = downLeft

		writeJSON(w, http.StatusOK, map[string]any{
			"settings": settings,
			"stats":    stats,
		})
	})

	mux.HandleFunc("/_control/reset", func(w http.ResponseWriter, r *http.Request) {
		c.mu.Lock()
		c.settings = Settings{}
		c.downUntil = time.Time{}
		c.stats = Stats{}
		c.mu.Unlock()

		writeJSON(w, http.StatusOK, map[string]string{"status": "reset"})
	})

	mux.HandleFunc("/_control", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
				"error": "use POST",
			})
			return
		}

		var settings Settings
		if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid json: " + err.Error(),
			})
			return
		}

		c.mu.Lock()
		c.settings = settings
		if settings.DownForS > 0 {
			c.downUntil = time.Now().Add(time.Duration(settings.DownForS) * time.Second)
		} else {
			c.downUntil = time.Time{}
		}
		c.mu.Unlock()

		writeJSON(w, http.StatusOK, map[string]any{"settings": settings})
	})
}

func (c *Controller) count(fn func(*Stats)) {
	c.mu.Lock()
	fn(&c.stats)
	c.mu.Unlock()
}

func hijackAndClose(w http.ResponseWriter) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		// Не смогли забрать соединение — отдаём хотя бы 503.
		// Хуже, чем разрыв, но лучше, чем тишина.
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "service unavailable",
		})
		return
	}

	conn, _, err := hijacker.Hijack()
	if err != nil {
		return
	}

	_ = conn.Close()
}

// WriteJSON — общий ответ моков. Экспортируется, потому что нужен
// самим обработчикам, а не только слою отказов.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	writeJSON(w, status, payload)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
