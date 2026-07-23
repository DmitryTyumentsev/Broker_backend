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
	BrokerID   string
	CustomerID string
	FixFor     string
	FixedBy    string
	FixationID uuid.UUID
	Status     Status
	FixedAt    time.Time
	ExpiresAt  time.Time
}
