package authhandlers

import (
	"Broker_backend/services/apigateway/internal/clients/authclient"
	"Broker_backend/services/apigateway/internal/transport/http/dto/authdto"
	"Broker_backend/services/apigateway/internal/transport/http/httperr"
	authv1 "Broker_backend/shared/pkg/grpc/gen/auth/v1"

	validate "github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type AuthHandler struct {
	logger     *zap.Logger
	authClient *authclient.Client
	validator  *validate.Validate
}

func NewAuthHandler(
	lg *zap.Logger,
	client *authclient.Client,
	vld *validate.Validate,
) *AuthHandler {
	return &AuthHandler{
		logger:     lg,
		authClient: client,
		validator:  vld,
	}
}

func (h *AuthHandler) Register(ctx *fiber.Ctx) error {
	var req authdto.RegisterRequest

	if err := ctx.BodyParser(&req); err != nil {
		return httperr.WriteBadRequest(ctx, "invalid request body")
	}

	if h.validator != nil {
		if err := h.validator.Struct(req); err != nil {
			return httperr.WriteBadRequest(ctx, "validation failed")
		}
	}

	grpcReq := httpToGRPCRegister(req) //хорошее ли решение каждый раз делать такую функцию отдельную или лучше прям тут парсить?

	resp, err := h.authClient.Register(ctx.Context(), grpcReq)
	if err != nil {
		return httperr.WriteGRPCError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(grpcToHTTPTokenPair(resp))
}

func (h *AuthHandler) Login(ctx *fiber.Ctx) error {
	var req authdto.LoginRequest

	if err := ctx.BodyParser(&req); err != nil {
		return httperr.WriteBadRequest(ctx, "invalid request body")
	}

	if h.validator != nil {
		if err := h.validator.Struct(req); err != nil {
			return httperr.WriteBadRequest(ctx, "validation failed")
		}
	}

	grpcReq := &authv1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
		DeviceId: req.DeviceID, //DeviceId откуда берем? когда и где запрашиваем?
	}

	resp, err := h.authClient.Login(ctx.Context(), grpcReq)
	if err != nil {
		return httperr.WriteGRPCError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(grpcToHTTPTokenPair(resp))
}

func (h *AuthHandler) Refresh(ctx *fiber.Ctx) error {
	var req authdto.RefreshRequest

	if err := ctx.BodyParser(&req); err != nil {
		return httperr.WriteBadRequest(ctx, "invalid request body")
	}

	if h.validator != nil {
		if err := h.validator.Struct(req); err != nil {
			return httperr.WriteBadRequest(ctx, "validation failed")
		}
	}

	grpcReq := &authv1.RefreshRequest{
		RefreshToken: req.RefreshToken,
		DeviceId:     req.DeviceID,
	}

	resp, err := h.authClient.Refresh(ctx.Context(), grpcReq)
	if err != nil {
		return httperr.WriteGRPCError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(grpcToHTTPTokenPair(resp))
}

func (h *AuthHandler) Logout(ctx *fiber.Ctx) error {
	var req authdto.LogoutRequest

	if err := ctx.BodyParser(&req); err != nil {
		return httperr.WriteBadRequest(ctx, "invalid request body")
	}

	if h.validator != nil {
		if err := h.validator.Struct(req); err != nil {
			return httperr.WriteBadRequest(ctx, "validation failed")
		}
	}

	grpcReq := &authv1.RefreshRequest{
		RefreshToken: req.RefreshToken,
		DeviceId:     req.DeviceID,
	}

	resp, err := h.authClient.Logout(ctx.Context(), grpcReq)
	if err != nil {
		return httperr.WriteGRPCError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(authdto.LogoutResponse{
		AllDevice: resp.AllDevice,
		DeviceID:  resp.DeviceId,
	})
}

func httpToGRPCRegister(req authdto.RegisterRequest) *authv1.RegisterRequest {
	return &authv1.RegisterRequest{
		Email:      req.Email,
		Password:   req.Password,
		LastName:   req.LastName,
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		DeviceId:   req.DeviceID,
	}
}

func grpcToHTTPTokenPair(resp *authv1.TokenPairResponse) authdto.TokenPairResponse {
	return authdto.TokenPairResponse{
		Access:       resp.AccessToken,
		Refresh:      resp.RefreshToken,
		ExpiresInSec: resp.ExpiresInSec,
	}
}
