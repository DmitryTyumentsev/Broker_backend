package domain

import (
	"context"
	"time"

	"Donate_backend/services/authservice/internal/domain/entity"
)

type UserRepository interface {
	Save(ctx context.Context, user entity.User) error
	FindByEmail(ctx context.Context, email string) (entity.User, error)
	FindByUsername(ctx context.Context, username string) (entity.User, error)
}

type RefreshSessionRepository interface {
	Save(ctx context.Context, session entity.RefreshSession) error
	Rotate(ctx context.Context, oldHash string, newSession entity.RefreshSession) error
	Revoke(ctx context.Context, hash string) error
	FindByHash(ctx context.Context, hash string) (entity.RefreshSession, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hashPass string, rawPass string) bool
}

type AccessTokenIssuer interface {
	Issue(userID string, deviceID string, role string, now time.Time) (string, error)
}

type RefreshTokenService interface {
	New() (string, error)
	Hash(token string) string
}

type Clock interface {
	Now() time.Time
}
