package usecases

import (
	"Donate_backend/services/authservice/internal/domain"
	"context"
)

type UserRepository interface {
	Save(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByUsername(ctx context.Context, username string) (domain.User, error)
}

type RefreshSessionRepository interface {
	Save(ctx context.Context, session domain.Session) error
	FindByRefresh(ctx context.Context, refreshHash string) (domain.Session, error)
	Rotate(ctx context.Context, oldRefreshHash string) (newRefreshHash string, err error)
	Revoke(ctx context.Context, refreshHash string) error
}

type AccessTokenIssuer interface {
	Issue(ctx context.Context, deviceID, userID, sessionID string) (accessToken string, err error)
}

type PasswordHasher interface {
	HashPassword(rawPassword string) (hashPassword string, err error)
	Verify(rawPassword, hashPassword string) bool
}

type RefreshTokenService interface {
	New() (rawRefresh string, err error)
	Hash(rawRefreshToken string) (refreshToken string, err error)
}
