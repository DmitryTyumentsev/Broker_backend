package entity

import "time"

type RefreshSession struct {
	RefreshTokenHash           string
	DeviceID                   string
	CreatedAt                  time.Time
	ExpiresAt                  time.Time
	RevokedAt                  *time.Time
	ReplacedByRefreshTokenHash *string
}
