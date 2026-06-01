package authhandlers

import (
	"errors"

	"Broker_backend/services/apigateway/internal/clients/authclient"
	"Broker_backend/services/apigateway/internal/transport/http/dto/authdto"
	"Broker_backend/services/apigateway/internal/transport/http/httperr"
	"Broker_backend/services/apigateway/internal/transport/http/middleware"
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
	if lg == nil {
		lg = zap.NewNop()
	}

	return &AuthHandler{
		logger:     lg,
		authClient: client,
		validator:  vld,
	}
}

func (h *AuthHandler) Validate() error {
	switch {
	case h == nil:
		return errors.New("auth handler is nil")
	case h.authClient == nil:
		return errors.New("auth client is required")
	default:
		return h.authClient.Validate()
	}
}

func (h *AuthHandler) Register(ctx *fiber.Ctx) error {
	var req authdto.RegisterRequest

	if err := ctx.BodyParser(&req); err != nil {
		return httperr.WriteBadRequest(ctx, "invalid request body")
	}

	if err := h.validate(req); err != nil {
		return httperr.WriteBadRequest(ctx, "validation failed")
	}

	resp, err := h.authClient.Register(ctx.UserContext(), httpToGRPCRegister(req))
	if err != nil {
		h.logger.Warn("register grpc failed", zap.Error(err))
		return httperr.WriteGRPCError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(grpcToHTTPTokenPair(resp))
}

func (h *AuthHandler) Login(ctx *fiber.Ctx) error {
	var req authdto.LoginRequest

	if err := ctx.BodyParser(&req); err != nil {
		return httperr.WriteBadRequest(ctx, "invalid request body")
	}

	if err := h.validate(req); err != nil {
		return httperr.WriteBadRequest(ctx, "validation failed")
	}

	grpcReq := &authv1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
		DeviceId: req.DeviceID,
	}

	resp, err := h.authClient.Login(ctx.UserContext(), grpcReq)
	if err != nil {
		h.logger.Warn("login grpc failed", zap.Error(err))
		return httperr.WriteGRPCError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(grpcToHTTPTokenPair(resp))
}

func (h *AuthHandler) Refresh(ctx *fiber.Ctx) error {
	var req authdto.RefreshRequest

	if err := ctx.BodyParser(&req); err != nil {
		return httperr.WriteBadRequest(ctx, "invalid request body")
	}

	if err := h.validate(req); err != nil {
		return httperr.WriteBadRequest(ctx, "validation failed")
	}

	grpcReq := &authv1.RefreshRequest{
		RefreshToken: req.RefreshToken,
		DeviceId:     req.DeviceID,
	}

	resp, err := h.authClient.Refresh(ctx.UserContext(), grpcReq)
	if err != nil {
		h.logger.Warn("refresh grpc failed", zap.Error(err))
		return httperr.WriteGRPCError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(grpcToHTTPTokenPair(resp))
}

func (h *AuthHandler) Logout(ctx *fiber.Ctx) error {
	var req authdto.LogoutRequest

	if err := ctx.BodyParser(&req); err != nil {
		return httperr.WriteBadRequest(ctx, "invalid request body")
	}

	if err := h.validate(req); err != nil {
		return httperr.WriteBadRequest(ctx, "validation failed")
	}

	grpcReq := &authv1.RefreshRequest{
		RefreshToken: req.RefreshToken,
		DeviceId:     req.DeviceID,
	}

	resp, err := h.authClient.Logout(ctx.UserContext(), grpcReq)
	if err != nil {
		h.logger.Warn("logout grpc failed", zap.Error(err))
		return httperr.WriteGRPCError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(authdto.LogoutResponse{
		AllDevice: resp.AllDevice,
		DeviceID:  resp.DeviceId,
	})
}

func (h *AuthHandler) Me(ctx *fiber.Ctx) error {
	claims, ok := middleware.CurrentClaims(ctx)
	if !ok {
		return httperr.WriteUnauthorized(ctx, "auth context is missing")
	}

	return ctx.Status(fiber.StatusOK).JSON(authdto.MeResponse{
		UserID:   claims.UserID,
		DeviceID: claims.DeviceID,
		Role:     claims.Role,
	})
}

func (h *AuthHandler) AdminPing(ctx *fiber.Ctx) error {
	claims, ok := middleware.CurrentClaims(ctx)
	if !ok {
		return httperr.WriteUnauthorized(ctx, "auth context is missing")
	}

	return ctx.Status(fiber.StatusOK).JSON(authdto.AdminPingResponse{
		OK:       true,
		UserID:   claims.UserID,
		DeviceID: claims.DeviceID,
		Role:     claims.Role,
	})
}

func (h *AuthHandler) validate(v any) error {
	if h.validator == nil {
		return nil
	}

	return h.validator.Struct(v)
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
	if resp == nil {
		return authdto.TokenPairResponse{}
	}

	return authdto.TokenPairResponse{
		Access:       resp.AccessToken,
		Refresh:      resp.RefreshToken,
		ExpiresInSec: resp.ExpiresInSec,
	}
}
