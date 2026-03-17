package httperr

import (
	"Donate_backend/services/apigateway/internal/http/dto"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/status"
)

func WriteGRPCError(c *fiber.Ctx, err error) error {
	resp := convertGRPCError(err)
	return c.Status(resp.Code).JSON(resp)
}

func convertGRPCError(err error) *dto.ErrorResponse {
	st := status.Convert(err)
	code := st.Code()
	message := st.Message()
	fields := st

	return &dto.ErrorResponse{
		Code:    code,
		Message: message,
		Fields:  fields,
	}
}
