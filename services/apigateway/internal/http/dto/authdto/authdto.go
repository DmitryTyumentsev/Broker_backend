package authdto

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Username string `json:"username" validate:"required,min=3,max=32"`
	DeviceID string `json:"device_id" validate:"required"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
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
