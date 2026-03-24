package grpchandler

import (
	"Donate_backend/services/authservice/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapError(err error) error {
	st := status.New(err) //TODO: зачем нужен status.New вообще?
	switch err {          //TODO: я же могу так написать? или нам нужно по типу сравнить? немного путаница - switch type, switch err и switch
	// и в case errors.Is. Видимо разница в том что где-то по значению а где-то по типу сравнение идет(предполагаю).
	//Разбери пж подробнее
	case domain.ErrUsernameExist:
		return st.Error(codes.BadRequest, "username already exist")
	case domain.ErrEmailExist:
		return st.Error(codes.BadRequest, "email already exist")
	case domain.ErrGeneral:
		return st.Error(codes.BadRequest, "something went wrong")
	}
	return err
}
