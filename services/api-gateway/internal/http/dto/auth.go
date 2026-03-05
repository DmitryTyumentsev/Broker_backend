package dto

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type LoginUserRequest struct {
	Identifier string `json:"Identifier"`
	Password   string `json:"password"`
}

type TokenPairResponse struct {
	Code         int32  `json:"code"`
	Message      string `json:"message"` //TODO: я должен 2 json отдавать или один? допустим при регистрации, надо вернуть все 4 поля. А при refresh последние 2. А при logout первые 2
	Access       string `json:"access"`
	ExpiresInSec int64  `json:"expires_in_sec"`
}
