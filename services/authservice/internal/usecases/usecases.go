package usecases

import (
	"Donate_backend/services/authservice/internal/config"
	"Donate_backend/services/authservice/internal/domain"

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
