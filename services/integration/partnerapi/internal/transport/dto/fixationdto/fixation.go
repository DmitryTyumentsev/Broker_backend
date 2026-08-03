package fixationdto

import (
	"time"

	"github.com/google/uuid"
)

type FixationRequest struct {
	Phone     string    `json:"phone" validate:"required"`
	ProjectID uuid.UUID `json:"project_id" validate:"required,uuid"`
}

type FixationResponse struct {
	FixationID uuid.UUID `json:"fixation_id"`
	FixedAt    time.Time `json:"fixed_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}
