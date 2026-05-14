package usecases

type RegisterRequest struct {
	Email      string
	Password   string
	MiddleName string
	DeviceID   string
}

type TokenPairResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresInSec int64
}
