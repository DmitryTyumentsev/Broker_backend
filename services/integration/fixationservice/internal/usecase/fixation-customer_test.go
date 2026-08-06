package usecase

import (
	"Broker_backend/services/integration/fixationservice/internal/config"
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

const (
	randomPhone = "8(999)999-99-99"
)

type mockServiceInterface interface {
	NewFixation(ctx context.Context, req *FixationRequest) (*entity.Fixation, error)
}

type mockRepo struct {
	isExistsProjectID           func(ctx context.Context, projectID uuid.UUID) (bool, error)
	isUserIDInAgencyID          func(ctx context.Context, agencyID, userID uuid.UUID) (bool, error)
	fixationCurrent             func(ctx context.Context, phoneHash string, projectID uuid.UUID) (*entity.Fixation, error)
	insertNewFixation           func(ctx context.Context, f entity.Fixation) error
	insertAudit                 func(ctx context.Context, f entity.Fixation) error
	insertOutbox                func(ctx context.Context, f entity.Fixation) error
	updateFixationStatusExpired func(ctx context.Context, statusExpired entity.Status, id uuid.UUID) error
	clock                       func() time.Time
	svc mockServiceInterface
}

//type mockService struct {
//	repo *mockRepo
//	cfg  *config.Config
//}

func TestNewFixation_NoActiveFixation_InsertFixationAuditOutbox(t *testing.T) {
	repo := &mockRepo{
		isExistsProjectID: func(context.Context, uuid.UUID) (bool, error) {
			return true, nil
		},
		isUserIDInAgencyID: func(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
			return true, nil
		},
		clock: func() time.Time {
			return time.Now()
		},
		fixationCurrent: func(context.Context, string, uuid.UUID) (*entity.Fixation, error) {
			return &entity.Fixation{
				Status: entity.StatusNoRows,
			}, nil
		},
		svc:
	}

	req := &FixationRequest{
		AgencyID:  uuid.New(),
		FixFor:    uuid.New(),
		FixBy:     uuid.New(),
		Phone:     randomPhone,
		ProjectID: uuid.New(),
	}
	expected, err := svc.NewFixation(context.Background(), req) //я не понимаю почему структура svc должна иметь тип *Service чтобы я мог вызвать NewFixation. это же метод у которого
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
