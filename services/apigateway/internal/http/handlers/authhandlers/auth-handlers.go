package authhandlers

import (
	"Donate_backend/services/apigateway/internal/clients/authclient"
	"Donate_backend/services/apigateway/internal/http/dto/authdto"
	"Donate_backend/services/apigateway/internal/http/httperr"
	authv1 "Donate_backend/shared/pkg/grpc/gen/auth/v1"

	validate "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type AuthHandler struct {
	logger     *zap.Logger
	authclient *authclient.Client
	validator  *validate.Validate
}

func NewAuthHandler(lg *zap.Logger, client *authclient.Client, validator *validate.Validate) *AuthHandler {
	return &AuthHandler{
		logger:     lg,
		authclient: client,
		validator:  validator}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req authdto.RegisterRequest

	err := c.BodyParser(&req)
	if err != nil {
		return WriteGRPCError(err)
	}
	if err = h.Validator.Struct(req); err != nil {
		return WriteGRPCError(err)
	}
	ctx := c.UserContext()
	grpcReq := httpToGRPCRegister(req)
	grpcResp, err := h.Client.Register(ctx, grpcReq)
	if err != nil {
		return httperr.WriteGrpcError(ctx, err)
	}

	resp := grpcToHTTPRegister(grpcResp)
	// —тут setRefreshCookie и дальше
	return c.Status(fiber.StatusCreate).JSON(authdto.TokenPairResponse{Access: grpc.Resp, ExpiresInSec: grpc.ExpiresInSec})
}

func httpToGRPCRegister(req authdto.RegisterRequest) *authv1.RegisterRequest {
	return &authv1.RegisterRequest{
		Email:    req.Email,
		Pass:     req.Pass,
		Username: req.Username,
		DeviceID: req.DeviceID,
	}
}

func grpcToHTTPRegister(resp *authv1.TokenPairResponse) *authdto.TokenPairResponse {
	return &authdto.TokenPairResponse{
		Access:       resp.Access,
		ExpiresInSec: resp.ExpiresInSec,
	}
}
