package usecase

import (
	"Broker_backend/services/app/authservice/internal/domain/entity"
	"context"
	"fmt"
	"strings"
	"time"
)

func (s *Service) createTokenPair(
	ctx context.Context,
	userID string,
	deviceID string,
	role string,
	now time.Time,
) (*TokenPairResponse, error) {
	const op = "usecase.createTokenPair"

	rawRefreshToken, err := s.refreshToken.New()
	if err != nil {
		return nil, fmt.Errorf("%s: create refresh token: %w", op, err)
	}

	refreshHash := s.refreshToken.Hash(rawRefreshToken)

	session := entity.RefreshSession{
		RefreshTokenHash: refreshHash,
		UserID:           userID,
		DeviceID:         strings.TrimSpace(deviceID),
		CreatedAt:        now,
		ExpiresAt:        now.Add(s.config.Business.LifetimeRefreshToken),
	}

	if err := s.sessions.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("%s: save refresh session: %w", op, err)
	}

	accessToken, err := s.accessIssuer.Issue(userID, deviceID, role, now)
	if err != nil {
		return nil, fmt.Errorf("%s: issue access token: %w", op, err)
	}

	return &TokenPairResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresInSec: int64(s.config.Business.LifetimeAccessToken.Seconds()),
	}, nil
}
