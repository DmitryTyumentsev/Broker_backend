package usecases

import "errors"

var (
	ErrUsernameRequired = errors.New("username is empty")
	ErrEmailRequired    = errors.New("email is empty")
	ErrPassRequired     = errors.New("pass is empty")
	ErrEmailInvalid     = errors.New("email invalid")
	ErrPassEasy         = errors.New("pass too much easy")
)
