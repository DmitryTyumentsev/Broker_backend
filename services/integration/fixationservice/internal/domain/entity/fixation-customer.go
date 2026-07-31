package entity

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusConverted Status = "converted"
	StatusExpired   Status = "expired"
	StatusRemoved   Status = "removed"
)

type FixationCustomer struct {
	AgencyID      uuid.UUID
	PhoneHash     string
	FixFor        uuid.UUID
	FixedBy       uuid.UUID
	FixationIDNew uuid.UUID //в dto у меня FixationIDNew и FixationIDOld. есть ли смысл также и в entity делить если поделил в dto и есть ли смысл вообще делать разделение?
	FixationIDOld uuid.UUID
	StatusActive  Status
	StatusExpired Status
	ProjectID     uuid.UUID
	FixedAt       time.Time
	ExpiresAt     time.Time
}
