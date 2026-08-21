package grpcerr

import (
	"Broker_backend/services/integration/partnerapi/internal/transport/dto"
	"Broker_backend/services/integration/partnerapi/internal/transport/http/httperr"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func WriteGRPCToHTTPError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return httperr.WriteInternal(c)
	}

	httpStatus := grpcCodeToHTTPStatus(st.Code())

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
