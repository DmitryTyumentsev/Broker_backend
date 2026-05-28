package usecases

import (
	"Broker_backend/services/authservice/internal/config"
	"Broker_backend/services/authservice/internal/domain"
	"fmt"

	"go.uber.org/zap"
)

type Service struct {
	config       *config.Config
	logger       *zap.Logger
	users        domain.UserRepository
	sessions     domain.RefreshSessionRepository
	passHasher   domain.PasswordHasher
	accessIssuer domain.AccessTokenIssuer
	refreshToken domain.RefreshTokenService
	clock        domain.Clock
}

func NewService(
	cfg *config.Config,
	logger *zap.Logger,
	users domain.UserRepository,
	sessions domain.RefreshSessionRepository,
	passHasher domain.PasswordHasher,
	accessIssuer domain.AccessTokenIssuer,
	refreshToken domain.RefreshTokenService,
	clock domain.Clock,
) *Service {
	return &Service{
		config:       cfg,
		logger:       logger,
		users:        users,
		sessions:     sessions,
		passHasher:   passHasher,
		accessIssuer: accessIssuer,
		refreshToken: refreshToken,
		clock:        clock,
	}
}

func (s *Service) ensureDeps() error {
	switch {
	case s.config == nil:
		return fmt.Errorf("config is nil")
	case s.users == nil:
		return fmt.Errorf("users repository is nil")
	case s.sessions == nil:
		return fmt.Errorf("sessions repository is nil")
	case s.passHasher == nil:
		return fmt.Errorf("password hasher is nil")
	case s.accessIssuer == nil:
		return fmt.Errorf("access token issuer is nil")
	case s.refreshToken == nil:
		return fmt.Errorf("refresh token service is nil")
	case s.clock == nil:
		return fmt.Errorf("clock is nil")
	default:
		return nil
	}
}
