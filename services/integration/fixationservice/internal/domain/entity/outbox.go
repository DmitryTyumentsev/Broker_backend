package entity

import (
	"time"

	"github.com/google/uuid"
)

type Outbox struct {
	ID          int
	ObjectID    uuid.UUID
	ObjectType  string
	EventType   string
	Payload     string
	CreatedAt   time.Time
	PublishedAt time.Time
}
