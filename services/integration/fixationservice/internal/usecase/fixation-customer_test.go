package usecase

import (
	"Broker_backend/services/integration/fixationservice/internal/config"
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"Broker_backend/services/integration/fixationservice/internal/repository/postgres"
	"Broker_backend/shared/pkg/clock"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	randomPhone = "8(999)999-99-99"
)

type mockRepo struct {
	isExistsProjectID           func(ctx context.Context, projectID uuid.UUID) (bool, error)
	isUserIDInAgencyID          func(ctx context.Context, agencyID, userID uuid.UUID) (bool, error)
	fixationCurrent             func(ctx context.Context, phoneHash string, projectID uuid.UUID) (*entity.Fixation, error)
	insertNewFixation           func(ctx context.Context, f entity.Fixation) error
	insertAudit                 func(ctx context.Context, f entity.Fixation) error
	insertOutbox                func(ctx context.Context, f entity.Fixation) error
	updateFixationStatusExpired func(ctx context.Context, statusExpired entity.Status, id uuid.UUID) error
}

func (m *mockRepo) IsExistsProjectID(ctx context.Context, projectID uuid.UUID) (bool, error) {
	return m.isExistsProjectID(ctx, projectID)
}

func (m *mockRepo) IsUserIDInAgencyID(ctx context.Context, agencyID, userID uuid.UUID) (bool, error) {
	return m.isUserIDInAgencyID(ctx, agencyID, userID)
}

func (m *mockRepo) FixationCurrent(ctx context.Context, phoneHash string, projectID uuid.UUID) (*entity.Fixation, error) {
	return m.fixationCurrent(ctx, phoneHash, projectID)
}

func (m *mockRepo) InsertNewFixation(ctx context.Context, f entity.Fixation) error {
	return m.insertNewFixation(ctx, f)
}

func (m *mockRepo) InsertAudit(ctx context.Context, f entity.Fixation) error {
	return m.insertAudit(ctx, f)
}

func (m *mockRepo) InsertOutbox(ctx context.Context, f entity.Fixation) error {
	return m.insertOutbox(ctx, f)
}

func (m *mockRepo) UpdateFixationStatusExpired(ctx context.Context, statusExpired entity.Status, id uuid.UUID) error {
	return m.updateFixationStatusExpired(ctx, statusExpired, id)
}

func TestNewFixation_NoActiveFixation_InsertFixationAuditOutbox(t *testing.T) {
	repo := &mockRepo{
		isExistsProjectID: func(context.Context, uuid.UUID) (bool, error) {
			return true, nil
		},
		isUserIDInAgencyID: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
			return true, nil
		},
		fixationCurrent: func(context.Context, string, uuid.UUID) (*entity.Fixation, error) {
			return &entity.Fixation{
				Status: entity.StatusNoRows,
			}, nil
		},
		insertNewFixation: func(context.Context, entity.Fixation) error {
			return nil
		},
		insertAudit: func(ctx context.Context, f entity.Fixation) error {
			return nil
		},
		insertOutbox: func(ctx context.Context, f entity.Fixation) error {
			return nil
		},
	}
	mockClock := &clock.RealClock{}
	mockTxManager := &postgres.TxManager{}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(cfg, &zap.Logger{}, mockClock, repo, mockTxManager)

	req := &FixationRequest{
		AgencyID:  uuid.New(),
		FixFor:    uuid.New(),
		FixBy:     uuid.New(),
		Phone:     randomPhone,
		ProjectID: uuid.New(),
	}
	got, err := svc.NewFixation(context.Background(), req)
	if err != nil {
		t.Errorf("NewFixation failed, %v", err)
	}
	if got == &entity.Fixation{
		AgencyID: req.AgencyID,
		PhoneHash: req.Phone,//как проверить PhoneHash
		FixFor:     req.FixFor,
		FixBy:      req.FixBy,
		FixationID: fixationID//аналогично как проверить?
		Status: entity.StatusNoRows,
		ProjectID:  req.ProjectID,
		FixedAt:    time.Time// времена как проверить какие?
		ExpiresAt  time.Time
	}

	//пишем какие сценарии могут быть:
	//позитивный(успешная новая фиксация),
	//позитивный(успешная фиксация быстрее воркера помечает старую фиксацию expired, создает новую),
	//негативный (фиксация уже создана и expires_at не истек)
	//негативный(req невалиден: project_id, fix_for, fix_by не существуют)
	//негативный(в agency_id нет юзеров fix_for, fix_by)
	//негативный(проверка транзакции - FixationCurrent выполнился, после этого соединение закрылось)
	//негативный(50 горутин одновременно пробуют сделать новую активную фиксацию)
	//готовим тестовые данные:
	//создаем мок структуры под каждый тест, внутри req/entity/pgerrcode
	//создаем в каждом тесте экземпляры моков, заполняем, вызываем оригинальный метод NewFixation передавая мок в ресивере и на вход метода

}
