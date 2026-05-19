package entity

import "time"

type Header struct {
	AccessTokenType string
	AccessTokenAlg  string
}

type Payload struct {
	UserID    string
	DeviceID  string
	Role      string
	TokenType string
	CreatedAt time.Time
	ExpiresAt time.Time
}
