package usecase

type RegisterRequest struct {
	Email       string
	RawPassword string
	LastName    string
	FirstName   string
	MiddleName  *string
	DeviceID    string
}

type LoginRequest struct {
	Email       string
	RawPassword string
	DeviceID    string
}

type RefreshRequest struct {
	RefreshToken string
	DeviceID     string
}

type LogoutRequest struct {
	RefreshToken string
	DeviceID     string
}

type TokenPairResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresInSec int64
}

type LogoutResponse struct {
	AllDevice bool
	DeviceID  string
}
