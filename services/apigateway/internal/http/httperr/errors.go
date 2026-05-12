package httperr

import (
	"Donate_backend/services/apigateway/internal/http/dto"
	"context"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func WriteGRPCError(c context.Context, err error) error {
	resp := convertGRPCError(err)
	return c.Status(resp.Code).JSON(resp)
}

func convertGRPCError(err error) *dto.ErrorResponse{
	st := status.Convert(err) //TODO: еще раз как эта функция работает? и зачем она. Мы же и так передаем status.Error ?
	// почему status.Error мы не можем использовать дальше и его надо преобразовать? code := toHTTPCode(st.Code() )
	//TODO: какие есть основные grpc коды? какие номера у них? field := toHTTPField(code, st.Details() )
	//TODO: лучше целиком класть st или code и st.Details() ? message := toHTTPMessage(code, field, st.Message() )
	//TODO: хорошо ли для частоты кода в этом кейсе вынести message как переменную, а не написать в структуре?
	return &dto.ErrorResponse{
		Code: code,
		Field: field,
		Message: message
	},
}
func toHTTPCode(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return fiber.StatusBadRequest
	case codes.AlreadyExists:
		return fiber.StatusConflict
	default: // дописать коды default:
		return fiber.StatusInternalServerError
	}
}

func toHTTPField(code codes.Code, det st.Details() ) string{ // TODO: st.Details() перевести тут на входе в то что возвращает
	switch code {
	case code.StatusInternalServerError, code.Status502, code.Status504
	return "" //TODO: если у нас 500ые, нам нужно просто вернуть код и message. Нужно ли возвращать в field "" ?
	}
	return det.GetField()
}

func toHTTPMessage(code codes.Code, field, message string) string{
	switch code {
	case fiber.StatusInternalServerError, fiber.Status502, fiber.Status504:
		return "Что-то пошло не так, попробуйте позже"
	case codes.InvalidArgument:
		return "Поле"
	case codes.AlreadyExists:
	}
	return det.GetField()
}
