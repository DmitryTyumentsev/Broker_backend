package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"Broker_backend/services/authservice/internal/domain"
	"Broker_backend/services/authservice/internal/domain/entity"

	"go.uber.org/zap"
)

func (s *Service) Refresh(ctx context.Context, req *RefreshRequest) (*TokenPairResponse, error) {
	const op = "usecases.Refresh"

	if err := s.ensureDeps(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := validateRefreshInput(req); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	oldHash := s.refreshToken.Hash(req.RefreshToken)

	oldSession, err := s.sessions.FindByHash(ctx, oldHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrUnauthenticated)
		}

		return nil, fmt.Errorf("%s: find refresh session: %w", op, err)
	}

	now := s.clock.Now()

	if oldSession.RevokedAt != nil {
		return nil, fmt.Errorf("%s: %w", op, domain.ErrSessionRevoked)
	}

	if oldSession.ExpiresAt.Before(now) {
		return nil, fmt.Errorf("%s: %w", op, domain.ErrSessionExpired)
	}

	if oldSession.DeviceID != req.DeviceID {
		return nil, fmt.Errorf("%s: %w", op, domain.ErrDeviceMismatch)
	}

	user, err := s.users.FindByID(ctx, oldSession.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: find user by id: %w", op, err)
	}

	newRawRefreshToken, err := s.refreshToken.New()
	if err != nil {
		return nil, fmt.Errorf("%s: create refresh token: %w", op, err)
	}

	newHash := s.refreshToken.Hash(newRawRefreshToken)

	newSession := entity.RefreshSession{
		RefreshTokenHash: newHash,
		UserID:           oldSession.UserID,
		DeviceID:         oldSession.DeviceID,
		CreatedAt:        now,
		ExpiresAt:        now.Add(s.config.Business.LifetimeRefreshToken),
	}

	if err := s.sessions.Rotate(ctx, oldHash, newSession); err != nil {
		return nil, fmt.Errorf("%s: rotate refresh session: %w", op, err)
	}

	accessToken, err := s.accessIssuer.Issue(user.ID, req.DeviceID, string(user.Role), now)
	if err != nil {
		return nil, fmt.Errorf("%s: issue access token: %w", op, err)
	}

	s.logger.Info("refresh token rotated", zap.String("user_id", user.ID))

	return &TokenPairResponse{
		AccessToken:  accessToken,
		RefreshToken: newRawRefreshToken,
		ExpiresInSec: int64(s.config.Business.LifetimeAccessToken.Seconds()),
	}, nil
}

func validateRefreshInput(req *RefreshRequest) error {
	if req == nil {
		return ErrRefreshRequired
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	req.DeviceID = strings.TrimSpace(req.DeviceID)

	if req.RefreshToken == "" {
		return ErrRefreshRequired
	}

	if req.DeviceID == "" {
		return ErrDeviceIDRequired
	}

	return nil
}
