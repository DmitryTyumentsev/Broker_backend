package httperr

import "errors"

var (
	ErrGeneral          = errors.New("Что-то пошло не так, попробуйте позже")
	ErrWrongCredentials = errors.New("Неверные логин, почта или пароль")
)
