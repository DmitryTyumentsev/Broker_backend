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
	StatusNoRows    Status = ""
)

type Fixation struct {
	AgencyID   uuid.UUID
	PhoneHash  string
	FixFor     uuid.UUID
	FixBy      uuid.UUID
	FixationID uuid.UUID
	Status     Status
	ProjectID  uuid.UUID
	FixedAt    time.Time
	ExpiresAt  time.Time
}
