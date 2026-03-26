package domain

import (
	"Donate_backend/services/authservice/internal/domain/entity"
	"context"
	"time"
)

type UserRepository interface {
	Save(ctx context.Context, user entity.User) error
	FindByEmail(ctx context.Context, email string) (entity.User, error)
	FindByUsername(ctx context.Context, username string) (entity.User, error)
}

type RefreshSessionRepository interface {
	Save(ctx context.Context, session entity.Session) error
	Rotate(ctx context.Context, oldRefresh string, newSess entity.RefreshSession, now time.Time) error
	RevokeByHash(ctx context.Context, refresh string, now time.Time) error
	FindByHash(ctx context.Context, refreshHash string) (entity.RefreshSession, error)
}

type AccessTokenIssuer interface {
	Issue(userID, deviceID string) (accessToken string, err error)
}

type PasswordHasher interface {
	Hash(rawPass string) (hashPass string, err error)
	Verify(hashPass, rawPass string) bool
}

type RefreshTokenService interface {
	New() (rawRefresh string, err error)
	Hash(rawRefresh string) (hashRefresh string, err error)
}

type Clock interface {
	Now() time.Time
}
