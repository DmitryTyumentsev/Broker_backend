package domain

import "errors"

var (
	ErrUsernameExist = errors.New("username already exist")
	ErrEmailExist    = errors.New("email already exist")
	ErrGeneral       = errors.New("something went wrong")
)
