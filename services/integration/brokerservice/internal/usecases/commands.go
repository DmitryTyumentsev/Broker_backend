package usecases

import "github.com/google/uuid"

type FixationRequest struct {
	BrokerID   uuid.UUID
	CustomerID uuid.UUID
	FixFor     uuid.UUID
	FixedBy    uuid.UUID
	ProjectID  uuid.UUID
}
