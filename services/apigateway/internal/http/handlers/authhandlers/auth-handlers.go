package authhandlers

import (
	"Donate_backend/services/apigateway/internal/clients/authclient"
	"Donate_backend/services/apigateway/internal/http/dto"
	"Donate_backend/services/apigateway/internal/http/httperr"
	"Donate_backend/services/authservice/internal/config"
	authv1 "Donate_backend/shared/pkg/grpc/gen/auth/v1"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	config     *config.Config
	logger     *zap.Logger
	authclient authclient.Client
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(httperr.ErrWrongCredentials.Error())
	}

	ctx := c.UserContext()
	grpcReq := convertHTTPToGRPCRegister(&req)
	grpcResp, err := h.authclient.Register(ctx, grpcReq)
	if err != nil {
	}

	resp := convertGRPCToHTTPTokenPairs(grpcResp)
	setCookie(resp)
	return c.JSON(resp)
}

func convertHTTPToGRPCRegister(req *dto.RegisterRequest) *authv1.RegisterRequest {
	return &authv1.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		Username: req.Username,
		DeviceId: req.DeviceID,
	}
}

func convertGRPCToHTTPTokenPairs(resp *authv1.TokenPairResponse) *dto.TokenPairResponse {
	return &dto.TokenPairResponse{
		Access:       resp.AccessToken,
		Refresh:      resp.RefreshToken,
		ExpiresInSec: resp.ExpiresInSec,
	}
}
