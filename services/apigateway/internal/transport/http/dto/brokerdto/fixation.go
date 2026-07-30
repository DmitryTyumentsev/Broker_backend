package brokerdto

import (
	"time"

	"github.com/google/uuid"
)

type FixationRequest struct { //верно ли на разные методы использовать одни и те же dto добавив поля для совместного использования? или если хоть одно поле отличается - делай новую dto?
	Phone     string    `json:"phone" validate:"required"`
	ProjectID uuid.UUID `json:"project_id" validate:"required,uuid"`
}

type FixationResponse struct {
	FixedAt   time.Time `json:"fixed_at" validate:"required"`
	ExpiresAt time.Time `json:"expires_at"`
}
