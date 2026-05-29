package entity

import "time"

type RefreshSession struct {
	SessionID                  int64
	RefreshTokenHash           string
	UserID                     string
	DeviceID                   string
	CreatedAt                  time.Time
	ExpiresAt                  time.Time
	RevokedAt                  *time.Time
	ReplacedByRefreshTokenHash *string
}
