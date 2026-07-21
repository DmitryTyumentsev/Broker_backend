package cmd

import (
	"time"

	"github.com/google/uuid"
)

type BrokerID string
type CustomerID string
type FixedBy string
type FixFor string

type FixationCustomerRequest struct {
	BrokerID   BrokerID
	CustomerID CustomerID //по какому правилу надо создавать отдельно тип для поля? ок ли везде пихать? где есть смысл их делать и а где нет?
	FixFor     FixFor
	FixedBy    FixedBy
}

type FixationCustomerResponse struct {
	FixationID uuid.UUID
	Status     string
	FixedAt    time.Time
	ExpiresAt  time.Time
}
