package authdto

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Username string `json:"username" validate:"required,min=8,max=72"`
	DeviceID string `json:"device_id" validate:"required,min=2,max=100"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required,min=2,max=255"`
	Password   string `json:"password" validate:"required,min=2,max=255"`
	DeviceID   string `json:"device_id" validate:"required"`
}

type TokenPairResponse struct {
	Access       string `json:"access"`
	Refresh      string `json:"refresh"`
	ExpiresInSec int64  `json:"expires_in_sec"`
}
