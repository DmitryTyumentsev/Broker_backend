package entity

import "time"

const (
	StatusActive    = "active" //верно же в папке entity держать константы в файле по фиче?
	StatusConverted = "converted"
	StatusExpired   = "expired"
	StatusRemoved   = "removed"
)

type BrokerID string
type CustomerID string
type FixedBy string
type FixFor string

type FixationCustomerRequest struct {
	BrokerID   BrokerID
	CustomerID CustomerID
	FixFor     FixFor
	FixedBy    FixedBy
}

type FixationCustomerResponse struct {
	FixationID string
	Status     string
	FixedAt    time.Time
	ExpiresAt  time.Time
}
