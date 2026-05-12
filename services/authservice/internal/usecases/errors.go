package usecases

import "errors"

var (
	ErrEmailRequired    = errors.New("email is required")
	ErrEmailInvalid     = errors.New("email is invalid")
	ErrPasswordRequired = errors.New("password is required")
	ErrPasswordWeak     = errors.New("password is weak")
	ErrUsernameRequired = errors.New("username is required")
	ErrUsernameInvalid  = errors.New("username is invalid")
	ErrDeviceIDRequired = errors.New("device_id is required")
)
