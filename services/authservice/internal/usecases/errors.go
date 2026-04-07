package usecases

import "errors"

var (
	ErrEmailRequired    = errors.New("email is empty")
	ErrEmailInvalid     = errors.New("email invalid")
	ErrPassRequired     = errors.New("pass is empty")
	ErrPassEasy         = errors.New("pass too much easy")
	ErrUsernameRequired = errors.New("username is empty")
	ErrUsernameInvalid  = errors.New("username invalid")
)
