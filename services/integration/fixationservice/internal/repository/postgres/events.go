package postgres

import (
	"Broker_backend/services/integration/fixationservice/internal/domain/entity"
	"Broker_backend/shared/pkg/authz"
	"Broker_backend/shared/pkg/requestctx"
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	fixationAggregateType    = "fixation"
	fixationCreatedAction    = "created"
	fixationCreatedEventType = "fixation.created"
)

// fixationState — единый снимок фиксации для аудита и события outbox.
// Один контракт не даёт двум записям одной транзакции незаметно разъехаться.
type fixationState struct {
	FixationID uuid.UUID     `json:"fixation_id"`
	AgencyID   uuid.UUID     `json:"agency_id"`
	FixBy      uuid.UUID     `json:"fix_by"`
	FixFor     uuid.UUID     `json:"fix_for"`
	ProjectID  uuid.UUID     `json:"project_id"`
	PhoneHash  string        `json:"phone_hash"`
	Status     entity.Status `json:"status"`
	FixedAt    time.Time     `json:"fixed_at"`
	ExpiresAt  time.Time     `json:"expires_at"`
}

func marshalFixationState(f entity.Fixation) (string, error) {
	state := fixationState{
		FixationID: f.FixationID,
		AgencyID:   f.AgencyID,
		FixBy:      f.FixBy,
		FixFor:     f.FixFor,
		ProjectID:  f.ProjectID,
		PhoneHash:  f.PhoneHash,
		Status:     f.Status,
		FixedAt:    f.FixedAt,
		ExpiresAt:  f.ExpiresAt,
	}

	raw, err := json.Marshal(state)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

type auditActor struct {
	UserID    any
	AgencyID  any
	Role      any
	RequestID any
	ClientIP  any
}

// auditActorFromContext принципиально не использует f.FixBy: значения из
// запроса не являются доказательством того, кто выполнил действие.
func auditActorFromContext(ctx context.Context) auditActor {
	actor := auditActor{}

	if principal, ok := authz.PrincipalFromContext(ctx); ok {
		actor.UserID = principal.UserID
		actor.AgencyID = principal.AgencyID
		actor.Role = principal.Role
	}

	if requestID, ok := requestctx.RequestIDFromContext(ctx); ok {
		actor.RequestID = requestID
	}
	if clientIP, ok := requestctx.ClientIPFromContext(ctx); ok {
		actor.ClientIP = clientIP
	}

	return actor
}
