package usecases

import (
	"Donate_backend/services/auth-service/internal/domain"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByPasswordHash(ctx context.Context, passHash string) (domain.User, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, refreshHash string) error
	Update(ctx context.Context, oldRefreshHash string) (RefreshHash string, err error)
}

type AccessTokenManager interface {
	NewAccessToken(ctx context.Context, deviceID, userID, sessionID string) (accessToken string, err error)
}

type HasherPassword interface {
	HashPassword(rawPassword string) (password string, err error)
}

type HasherToken interface {
	HashRefreshToken(rawRefreshToken string) (refreshToken string, err error)
	HashAccessToken(rawAccessToken string) (accessToken string, err error)
}
