package usecases

import (
	"Broker_backend/services/brokerservice/internal/domain/interfaces"
	"Broker_backend/services/brokerservice/internal/infra/repositories/postgres"
	"fmt"

	"Broker_backend/services/brokerservice/internal/config"

	"go.uber.org/zap"
)

// ты написал что тут нужен интерфейс так как база потребитель. я видимо не совсем понимаю что означает потребитель. в моей картине потребитель - тот кто потребляет, запрашивает данные. есть юзкейс который вызывает метод в постгре. потребитель юзкейс который вызывает постгрю или потребитель постгря которая получает данные от юзкейса? они же друг для друга оба потребители - друг от друга получают данные, друг другу отдают их. или надо ставить интерфейс если хотим пойти в слой ниже? а потребитель причем тут?
type Service struct {
	config *config.Config
	logger *zap.Logger
	clock  interfaces.Clock
	pg     *postgres.Postgres //в продолжение вопроса про интерфейс из этого файла выше - тут выходит интерфейс должен быть? чтобы подменять бд. зачем он если да? зачем на FixationRepository интерфейс?
}

func NewService(
	cfg *config.Config,
	logger *zap.Logger,
	clock interfaces.Clock,
	pg *postgres.Postgres,
) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Service{
		config: cfg,
		logger: logger,
		clock:  clock,
		pg:     pg,
	}
}

func (s *Service) ensureDeps() error {
	switch {
	case s == nil:
		return fmt.Errorf("service is nil")
	case s.config == nil:
		return fmt.Errorf("config is nil")
	case s.clock == nil:
		return fmt.Errorf("clock is nil")
	case s.pg == nil:
		return fmt.Errorf("postgres is nil")
	default:
		return nil
	}
}
