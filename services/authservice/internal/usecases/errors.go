package usecases

import "errors"

var (
	ErrEmailRequired     = errors.New("email is required")
	ErrEmailInvalid      = errors.New("email is invalid")
	ErrPasswordRequired  = errors.New("password is required")
	ErrPasswordWeak      = errors.New("password is weak")
	ErrUsernameRequired  = errors.New("username is required")
	ErrDeviceIDRequired  = errors.New("device_id is required")
	ErrFirstNameRequired = errors.New("first_name is required")
	ErrLastNameRequired  = errors.New("last_name is required")
)
