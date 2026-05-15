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
