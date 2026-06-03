package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"Broker_backend/services/authservice/internal/domain"

	"go.uber.org/zap"
)

func (s *Service) Logout(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error) {
	const op = "usecases.Logout"

	if err := s.ensureDeps(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := validateLogoutInput(req); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	hash := s.refreshToken.Hash(req.RefreshToken)

	session, err := s.sessions.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return &LogoutResponse{
				AllDevice: false,
				DeviceID:  req.DeviceID,
			}, nil
		}

		return nil, fmt.Errorf("%s: find refresh session: %w", op, err)
	}

	if session.DeviceID != req.DeviceID {
		return nil, fmt.Errorf("%s: %w", op, domain.ErrDeviceMismatch)
	}

	if err := s.sessions.Revoke(ctx, hash); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("%s: revoke refresh session: %w", op, err)
		}
	}

	s.logger.Info("user logged out", zap.String("user_id", session.UserID))

	return &LogoutResponse{
		AllDevice: false,
		DeviceID:  req.DeviceID,
	}, nil
}

func validateLogoutInput(req *LogoutRequest) error {
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
