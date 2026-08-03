package usecase

import "github.com/google/uuid"

type FixationRequest struct {
	AgencyID  uuid.UUID
	FixFor    uuid.UUID
	Phone     string
	ProjectID uuid.UUID
}
