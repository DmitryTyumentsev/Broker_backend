package authdto

import "github.com/google/uuid"

type RegisterRequest struct {
	Email      string  `json:"email" validate:"required,email"`
	Password   string  `json:"password" validate:"required,min=8"`
	LastName   string  `json:"last_name" validate:"required,min=2,max=64"`
	FirstName  string  `json:"first_name" validate:"required,min=2,max=64"`
	MiddleName *string `json:"middle_name,omitempty" validate:"omitempty,min=2,max=64"`
	DeviceID   string  `json:"device_id" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	DeviceID string `json:"device_id" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	DeviceID     string `json:"device_id" validate:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	DeviceID     string `json:"device_id" validate:"required"`
}

type TokenPairResponse struct {
	Access       string `json:"access"`
	Refresh      string `json:"refresh"`
	ExpiresInSec int64  `json:"expires_in_sec"`
}

type LogoutResponse struct {
	AllDevice bool   `json:"all_device"`
	DeviceID  string `json:"device_id"`
}

// uuid.UUID сериализуется в JSON как строка (MarshalText),
// отдельного маппинга в string не нужно.
type MeResponse struct {
	AgencyID uuid.UUID `json:"agency_id"`
	UserID   uuid.UUID `json:"user_id"`
	DeviceID string    `json:"device_id"`
	Role     string    `json:"role"`
}

type AdminPingResponse struct {
	OK       bool      `json:"ok"`
	AgencyID uuid.UUID `json:"agency_id"`
	UserID   uuid.UUID `json:"user_id"`
	DeviceID string    `json:"device_id"`
	Role     string    `json:"role"`
}
