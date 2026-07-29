package httperr

import (
	"Broker_backend/services/apigateway/internal/transport/http/dto"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func WriteGRPCError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return WriteInternal(c)
	}

	httpStatus := grpcCodeToHTTPStatus(st.Code())

	return c.Status(httpStatus).JSON(dto.ErrorResponse{
		Code:    httpStatus,
		Message: st.Message(),
	})
}

func WriteHTTPError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	code, ok := status.FromError(err) //надо для http такое из ctx вытащить
	if !ok {
		return WriteInternal(c)
	}

	switch code {
	case fiber.StatusConflict:
		return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
			Code:    fiber.StatusConflict,
			Message: st.Message(),
		})
	}

	return c.Status(httpStatus).JSON(dto.ErrorResponse{
		Code:    httpStatus,
		Message: st.Message(),
	})
}

func grpcCodeToHTTPStatus(code codes.Code) int {
	switch code {
	case codes.InvalidArgument:
		return fiber.StatusBadRequest
	case codes.NotFound:
		return fiber.StatusNotFound
	case codes.AlreadyExists:
		return fiber.StatusConflict
	case codes.Unauthenticated:
		return fiber.StatusUnauthorized
	case codes.PermissionDenied:
		return fiber.StatusForbidden
	case codes.Unimplemented:
		return fiber.StatusNotImplemented
	case codes.Unavailable:
		return fiber.StatusServiceUnavailable
	case codes.DeadlineExceeded:
		return fiber.StatusGatewayTimeout
	case codes.Canceled:
		return fiber.StatusRequestTimeout
	default:
		return fiber.StatusInternalServerError
	}
}

func WriteBadRequest(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
		Code:    fiber.StatusBadRequest,
		Message: message,
	})
}

func WriteConflict(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
		Code:    fiber.StatusConflict,
		Message: message,
	})
}

func WriteInternal(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
		Code:    fiber.StatusInternalServerError,
		Message: "internal server error",
	})
}

func WriteUnauthorized(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
		Code:    fiber.StatusUnauthorized,
		Message: message,
	})
}

func WriteForbidden(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
		Code:    fiber.StatusForbidden,
		Message: message,
	})
}

func WritePayloadTooLarge(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusRequestEntityTooLarge).JSON(dto.ErrorResponse{
		Code:    fiber.StatusRequestEntityTooLarge,
		Message: message,
	})
}

func WriteTooManyRequests(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(dto.ErrorResponse{
		Code:    fiber.StatusTooManyRequests,
		Message: message,
	})
}

func WriteGatewayTimeout(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusGatewayTimeout).JSON(dto.ErrorResponse{
		Code:    fiber.StatusGatewayTimeout,
		Message: message,
	})
}

func WriteServiceUnavailable(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusServiceUnavailable).JSON(dto.ErrorResponse{
		Code:    fiber.StatusServiceUnavailable,
		Message: message,
	})
}
