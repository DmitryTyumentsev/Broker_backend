package domain

import "errors"

var (
	ErrUsernameExist             = errors.New("username already exists")
	ErrEmailExist                = errors.New("email already exists")
	ErrNotUnique                 = errors.New("not unique")
	ErrMustBeNotNull             = errors.New("must be not null")
	ErrBadRequest                = errors.New("bad request")
	ErrNotFound                  = errors.New("not found")
	ErrUnauthenticated           = errors.New("unauthenticated")
	ErrForbidden                 = errors.New("forbidden")
	ErrGeneral                   = errors.New("internal error")
	ErrNotUniqueEmail            = errors.New("email not unique")
	ErrUserRoleInvalid           = errors.New("user role invalid")
	ErrNotUniqueRefreshTokenHash = errors.New("refresh token hash not unique")
)
