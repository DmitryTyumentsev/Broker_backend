package handlers

import (
	"Donate_backend/shared/pkg/config"
)

type Handlers struct {
	Auth *AuthService
}

type AuthService struct {
	Config config.Loader
	//log    *log.Logger
}
