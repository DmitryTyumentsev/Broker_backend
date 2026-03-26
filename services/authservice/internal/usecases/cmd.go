package usecases

type RegisterRequest struct {
	Email, Password, Username, DeviceID string
}

type TokenPairResponse struct {
	AccessToken, RefreshToken string
	ExpiresInSec              int64
}
