package usecases

import (
	"Broker_backend/services/integration/brokerservice/internal/config"
	"Broker_backend/services/integration/brokerservice/internal/domain/entity"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ты написал что тут нужен интерфейс так как база потребитель. я видимо не совсем понимаю что означает потребитель. в моей картине потребитель - тот кто потребляет, запрашивает данные. есть юзкейс который вызывает метод в постгре. потребитель юзкейс который вызывает постгрю или потребитель постгря которая получает данные от юзкейса? они же друг для друга оба потребители - друг от друга получают данные, друг другу отдают их. или надо ставить интерфейс если хотим пойти в слой ниже? а потребитель причем тут?
type FixationRepository interface {
	Insert(ctx context.Context, fixationCustomer entity.FixationCustomer) error
	Update(ctx context.Context, fixationCustomer entity.FixationCustomer) error
}

type TxManager interface { //зачем нужен здесь этот интерфейс - не понимаю и неясно есть ли смысл особо в это вникать(есть ли смысл сильно внимание акцентировать на обертке и тому как она устроена)
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type Clock interface {
	Now() time.Time
}

type Service struct {
	cfg       *config.Config
	logger    *zap.Logger
	clock     Clock
	fixations FixationRepository
	tx        TxManager
}

//type ManagerRepository interface {
//	RoleByManagerID(managerID domain.ManagerID) (string, error)
//	FullNameByManagerID(managerID domain.ManagerID) (string, error)
//}
//type FixationCustomerRepository interface {
//	SaveActiveFixationCustomer(uuid string, expiresAT, fixedAt *time.Time, statusActive string, brokerID entity.BrokerID, fixedBy entity.FixedBy, fixFor entity.FixFor, customerID entity.CustomerID) error
//	UpdateFixationCustomer(brokerID entity.BrokerID, managerID entity.ManagerID, customerID entity.CustomerID) error
//	FreeFixationCustomer(brokerID entity.BrokerID, managerID entity.ManagerID) error
//	SelectStatusByCustomerID(customerID entity.CustomerID) (string, error) //под вопросом надо ли
//}
//
//type FixationCustomerService interface { //принято ли делить интерфейсы в сервисном уровне отдельно на Issue, Check и тд или просто когда фичю пилишь - создаешь интерфейс под сервис и под репо?
//	CreateActiveFixationCustomer(brokerID entity.BrokerID, fixedBy entity.FixedBy, fixFor entity.FixFor, customerID entity.CustomerID) error
//	CheckStatusCustomer(customerID entity.CustomerID) (string, error)
//}

func NewService(
	cfg *config.Config,
	logger *zap.Logger,
	clock Clock,
	fixation FixationRepository,
	tx TxManager,
) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Service{
		cfg:       cfg,
		logger:    logger,
		clock:     clock,
		fixations: fixation,
		tx:        tx,
	}
}

func (s *Service) ensureDeps() error {
	switch {
	case s == nil:
		return fmt.Errorf("service is nil")
	case s.cfg == nil:
		return fmt.Errorf("config is nil")
	case s.clock == nil:
		return fmt.Errorf("clock is nil")
	case s.fixations == nil:
		return fmt.Errorf("postgres is nil")
	case s.tx == nil:
		return fmt.Errorf("tx is nil")
	default:
		return nil
	}
}
