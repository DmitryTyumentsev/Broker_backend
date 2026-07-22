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

type FixationCustomerRequest struct {
	BrokerID   BrokerID
	CustomerID CustomerID
	FixFor     FixFor
	FixedBy    FixedBy
}

type FixationCustomerResponse struct {
	FixationID uuid.UUID
	Status     Status
	FixedAt    FixedAt
	ExpiresAt  ExpiresAt
}
