package usecases

import (
	"context"
	"fmt"
	"strings"

	"Donate_backend/services/authservice/internal/domain/entity"
	"Donate_backend/services/authservice/internal/pkg/validators"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*TokenPairResponse, error) {
	const op = "usecases.Register"

	if s == nil {
		return nil, fmt.Errorf("%s: service is nil", op)
	}

	if err := validateRegisterInput(req); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := s.ensureRegisterDeps(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	passwordHash, err := s.passHasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("%s: hash password: %w", op, err)
	}

	now := s.clock.Now()

	user := entity.User{
		ID:        uuid.NewString(),
		Email:     strings.TrimSpace(req.Email),
		Username:  strings.TrimSpace(req.Username),
		PassHash:  passwordHash,
		Role:      "user",
		CreatedAt: now,
	}

	if err := s.users.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("%s: save user: %w", op, err)
	}

	rawRefreshToken, err := s.refreshToken.New()
	if err != nil {
		return nil, fmt.Errorf("%s: create refresh token: %w", op, err)
	}

	refreshHash := s.refreshToken.Hash(rawRefreshToken)

	session := entity.RefreshSession{
		Hash:      refreshHash,
		UserID:    user.ID,
		DeviceID:  req.DeviceID,
		CreatedAt: now,
		ExpiresAt: now.Add(s.config.Business.LifetimeRefreshToken),
	}

	if err := s.sessions.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("%s: save refresh session: %w", op, err)
	}

	accessToken, err := s.accessIssuer.Issue(user.ID, req.DeviceID, user.Role, now)
	if err != nil {
		return nil, fmt.Errorf("%s: issue access token: %w", op, err)
	}

	if s.logger != nil {
		s.logger.Info("user registered", zap.String("user_id", user.ID))
	}

	return &TokenPairResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresInSec: int64(s.config.Business.LifetimeAccessToken.Seconds()),
	}, nil
}

func validateRegisterInput(req *RegisterRequest) error {
	if req == nil {
		return ErrEmailRequired
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)

	if req.Email == "" {
		return ErrEmailRequired
	}

	if !validators.IsValidEmail(req.Email) {
		return ErrEmailInvalid
	}

	if req.Password == "" {
		return ErrPasswordRequired
	}

	if !validators.IsStrongPassword(req.Password) {
		return ErrPasswordWeak
	}

	if req.Username == "" {
		return ErrUsernameRequired
	}

	if !validators.IsValidUsername(req.Username) {
		return ErrUsernameInvalid
	}

	if req.DeviceID == "" {
		return ErrDeviceIDRequired
	}

	return nil
}

func (s *Service) ensureRegisterDeps() error {
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
