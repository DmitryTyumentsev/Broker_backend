package domain

import (
	entity2 "Broker_backend/services/app/authservice/internal/domain/entity"
	"context"
	"time"
)

type UserRepository interface {
	Save(ctx context.Context, user entity2.User) error
	FindByID(ctx context.Context, id string) (entity2.User, error)
	FindByEmail(ctx context.Context, email string) (entity2.User, error)
}

type RefreshSessionRepository interface {
	Save(ctx context.Context, session entity2.RefreshSession) error
	Rotate(ctx context.Context, oldHash string, newSession entity2.RefreshSession) error
	Revoke(ctx context.Context, hash string) error
	FindByHash(ctx context.Context, hash string) (entity2.RefreshSession, error)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hashPass string, rawPass string) bool
}

type AccessTokenIssuer interface {
	Issue(agencyID string, userID string, deviceID string, role string, now time.Time) (string, error)
}

type RefreshTokenService interface {
	New() (string, error)
	Hash(token string) string
}

type Clock interface {
	Now() time.Time
}
