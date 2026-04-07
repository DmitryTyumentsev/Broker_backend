package entity

import "time"

type RefreshSession struct {
	Hash           string
	UserID         string
	DeviceID       string
	CreatedAt      time.Time
	ExpiresAt      time.Time
	RevokedAt      *time.Time
	ReplacedByHash *string
}
