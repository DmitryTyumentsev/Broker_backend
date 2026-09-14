package amocrm

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

var (
	ErrTemporary          = errors.New("temporary error")
	ErrTimeout            = errors.New("timeout error")
	ErrBadRequest         = errors.New("bad request")
	ErrPermanent          = errors.New("permanent error")
	ErrSomethingWentWrong = errors.New("something went wrong")
)

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP error %d: %s", e.StatusCode, e.Body)
}

func MapError(err error) error {
	if err == nil {
		return nil
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) { //не понимаю каким образом ошибка будет иметь тип HTTPError. это же просто структура которую мы завели, почему у error будет именно этот тип вообще не понял
		return err
	}
	switch httpErr.StatusCode {
	case fiber.StatusBadRequest:
		return ErrBadRequest
	case fiber.StatusInternalServerError:
		return ErrPermanent
	default:
		return ErrSomethingWentWrong
	}
}
