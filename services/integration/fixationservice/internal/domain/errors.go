package domain

import "errors"

var (
	ErrNotUnique            = errors.New("not unique")
	ErrMustBeNotNull        = errors.New("must be not null")
	ErrBadRequest           = errors.New("bad request")
	ErrNotFound             = errors.New("not found")
	ErrUnauthenticated      = errors.New("unauthenticated")
	ErrForbidden            = errors.New("forbidden")
	ErrGeneral              = errors.New("internal error")
	ErrFixationAlreadyExist = errors.New("fixation already exist")
)
