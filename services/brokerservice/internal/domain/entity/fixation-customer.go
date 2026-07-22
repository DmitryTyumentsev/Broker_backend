package entity

import (
	"time"

	"github.com/google/uuid"
)

type Status string

type BrokerID string
type CustomerID string
type FixedBy string
type FixFor string

type FixedAt time.Time
type ExpiresAt time.Time

const (
	StatusActive    Status = "active"
	StatusConverted Status = "converted"
	StatusExpired   Status = "expired"
	StatusRemoved   Status = "removed"
)

type FixationCustomer struct {
	BrokerID   BrokerID
	CustomerID CustomerID
	FixFor     FixFor
	FixedBy    FixedBy
	FixationID uuid.UUID
	Status     Status
	FixedAt    FixedAt
	ExpiresAt  ExpiresAt
}
