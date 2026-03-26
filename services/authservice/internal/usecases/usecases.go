package usecases

import (
	"Donate_backend/services/authservice/internal/config"
	"Donate_backend/services/authservice/internal/domain"
)

type Service struct {
	userRepo     domain.UserRepository
	refSessRepo  domain.RefreshSessionRepository
	accessIssier domain.AccessTokenIssuer
	passHasher   domain.PasswordHasher
	refService   domain.RefreshTokenService
	clock        domain.Clock
	config       *config.Config
	logger       *zap.Logger
}

func NewService(userRepo domain.UserRepository, refSessRepo domain.RefreshSessionRepository, accessIssier domain.AccessTokenIssuer,
	passHasher domain.PasswordHasher, refService domain.RefreshTokenService, clock domain.Clock, config *config.Config, logger *zap.Logger) *Service {
	return &Service{
		userRepo:     userRepo,
		refSessRepo:  refSessRepo,
		accessIssier: accessIssier,
		passHasher:   passHasher,
		refService:   refService,
		clock:        clock,
		config:       config,
		logger:       logger,
	}
}
