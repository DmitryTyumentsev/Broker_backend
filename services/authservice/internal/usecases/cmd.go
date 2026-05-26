package usecases

type RegisterRequest struct {
	Email       string
	RawPassword string
	LastName    string
	FirstName   string
	MiddleName  *string
	DeviceID    string
}

type TokenPairResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresInSec int64
}

type AuthRequest struct {
	Email       string
	RawPassword string
	DeviceID    string
}

type RefreshRequest struct {
	RefreshToken string
	DeviceID     string
}

type LogoutRequest struct {
	AllDevice bool
	DeviceID  string
}
