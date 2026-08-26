package integrationtest

import (
	"Broker_backend/services/integration/fixationservice/internal/config"
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"Broker_backend/services/integration/fixationservice/internal/usecase"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

//const (
//	goroutines       = 50
//	testPhone        = "8(999)111-22-33"
//	fixationDuration = 60 * 24 * time.Hour
//	hashSecret       = "integration-test-secret"
//)
//
//// realClock — настоящие часы. В интеграционном тесте подменять время
//// незачем: проверяется поведение базы, а не арифметика сроков.
//type realClock struct{}
//
//func (realClock) Now() time.Time { return time.Now().UTC() }
//
//func TestNewFixation_50Goroutines_ExactlyOneSucceeds(t *testing.T) {
//	agencyID, projectID, userID := seedAgencyProjectUser(t)
//
//	// ── Настоящий стек целиком. Ни одного мока ──────────────────────
//	txManager := postgres.NewTxManager(testPool)
//	repo := postgres.NewRepository(txManager)
//
//	cfg := &config.Config{}
//	cfg.Business.FixationDuration = fixationDuration
//	cfg.Business.HashSecret = hashSecret
//
//	svc := usecase.NewService(cfg, zap.NewNop(), realClock{}, repo, txManager)
//
//	req := &usecase.FixationRequest{
//		AgencyID:  agencyID,
//		FixFor:    userID,
//		FixBy:     userID,
//		ProjectID: projectID,
//		Phone:     testPhone,
//	}
//
//	// ── Запуск ──────────────────────────────────────────────────────
//	//
//	// start — стартовый барьер. Все горутины блокируются на чтении из
//	// незакрытого канала. close(start) отпускает их ОДНОВРЕМЕННО:
//	// чтение из закрытого канала возвращается сразу и у всех.
//	//
//	// Без барьера первые горутины успеют отработать раньше, чем
//	// запустятся последние. Гонки не случится, и тест позеленеет
//	// даже на сломанном индексе.
//	start := make(chan struct{})
//
//	// Буфер на goroutines — иначе горутины залипнут на отправке,
//	// пока основная не начнёт читать, и Wait никогда не вернётся.
//	results := make(chan error, goroutines)
//
//	var wg sync.WaitGroup
//	for i := 0; i < goroutines; i++ {
//		wg.Add(1)
//		go func() {
//			defer wg.Done()
//			<-start // ждём здесь
//			_, err := svc.NewFixation(context.Background(), req)
//			results <- err
//		}()
//	}
//
//	close(start)
//	wg.Wait() // без него тест завершится раньше горутин
//	close(results)
//
//	// ── Разбор ──────────────────────────────────────────────────────
//	//
//	// t.Errorf зовём только здесь, в основной горутине. Из горутин
//	// после завершения теста это даёт панику всего прогона.
//	var success, conflict int
//	var unexpected []error
//
//	for err := range results {
//		switch {
//		case err == nil:
//			success++
//		case errors.Is(err, domain.ErrNotUnique),
//			errors.Is(err, domain.ErrFixationAlreadyExist):
//			conflict++
//		default:
//			unexpected = append(unexpected, err)
//		}
//	}
//
//	// Проверяем КОЛИЧЕСТВО, а не «победила нулевая горутина»:
//	// порядок недетерминирован, выиграть может любая.
//	if success != 1 {
//		t.Errorf("успехов: %d, ожидали ровно 1 (конфликтов %d)", success, conflict)
//	}
//	if conflict != goroutines-1 {
//		t.Errorf("конфликтов: %d, ожидали %d", conflict, goroutines-1)
//	}
//	for _, err := range unexpected {
//		t.Errorf("неожиданная ошибка вместо конфликта: %v", err)
//	}
//
//	// Контрольная проверка в самой базе. Главная в этом тесте:
//	// выше проверялось, что вернул код, здесь — что реально записано.
//	phoneHash := usecase.HashPhone(testPhone, hashSecret) // ← подставь свою функцию
//	if n := countActiveFixations(t, phoneHash, projectID); n != 1 {
//		t.Fatalf("активных фиксаций в базе: %d, ожидали 1", n)
//	}
//}
//
//// Второй сценарий, который тоже проверяется только на живой базе:
//// протухшая фиксация не должна мешать создать новую.
//func TestNewFixation_ExpiredFixation_DoesNotBlockNew(t *testing.T) {
//	agencyID, projectID, userID := seedAgencyProjectUser(t)
//
//	txManager := postgres.NewTxManager(testPool)
//	repo := postgres.NewRepository(txManager)
//
//	cfg := &config.Config{}
//	cfg.Business.FixationDuration = fixationDuration
//	cfg.Business.HashSecret = hashSecret
//
//	svc := usecase.NewService(cfg, zap.NewNop(), realClock{}, repo, txManager)
//
//	req := &usecase.FixationRequest{
//		AgencyID:  agencyID,
//		FixFor:    userID,
//		FixBy:     userID,
//		ProjectID: projectID,
//		Phone:     "8(999)444-55-66",
//	}
//
//	if _, err := svc.NewFixation(context.Background(), req); err != nil {
//		t.Fatalf("первая фиксация: %v", err)
//	}
//
//	// Сдвигаем срок в прошлое НАПРЯМУЮ в базе, не запуская никакой воркер.
//	// Смысл: горячий путь обязан работать, даже если фоновый процесс лежит.
//	_, err := testPool.Exec(context.Background(),
//		`UPDATE integration.fixations SET expires_at = now() - interval '1 day'
//		 WHERE project_id = $1`, projectID)
//	if err != nil {
//		t.Fatalf("сдвиг срока: %v", err)
//	}
//
//	if _, err := svc.NewFixation(context.Background(), req); err != nil {
//		t.Fatalf("вторая фиксация после протухания: %v", err)
//	}
//}

func TestNewFixation_RaceFixations_ActiveFixationNoRows_OneOfFixationsSuccess(t *testing.T) {
	cfg := &config.Config{
		Business: config.BusinessConfig{
			HashSecret:       "text-text-text-text",
			FixationDuration: 24 * 30 * time.Hour,
		},
	}
	repo := usecase.FixationRepository
	tx := usecase.TxManager
	now := time.Now().UTC()
	projectID, err := uuid.Parse("text-text-text-text")
	if err != nil {
		t.Fatal(err)
	}
	req := &usecase.FixationRequest{
		AgencyID:  uuid.New(),
		FixFor:    uuid.New(),
		FixBy:     uuid.New(),
		Phone:     "+7(999)999-99-99",
		ProjectID: projectID,
	}
	svc := usecase.NewService(cfg, zap.NewNop(), now, repo, tx)
	chErr := make(chan error)
	chRes := make(chan *entity.Fixation)
	go func() {
		res, err := svc.NewFixation(context.Background(), req)
		chErr <- err
		chRes <- res
	}()
	res := <-chRes
	if res == nil {
		t.Fatal("res is nil")
	}
	err = <-chErr
	if err != nil {
		t.Fatal(err)
	}
}
