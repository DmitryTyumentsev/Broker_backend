package fixationdto

import (
	"time"

	"github.com/google/uuid"
)

type FixationRequest struct {
	Phone     string    `json:"phone" validate:"required"`
	ProjectID uuid.UUID `json:"project_id" validate:"required,uuid"`
	FixFor    uuid.UUID `json:"fix_for" validate:"omitempty,uuid"`
}

type Meta struct {
	AgencyID uuid.UUID
	FixBy    uuid.UUID
}

type FixationResponse struct {
	FixationID uuid.UUID `json:"fixation_id"`
	FixedAt    time.Time `json:"fixed_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}
