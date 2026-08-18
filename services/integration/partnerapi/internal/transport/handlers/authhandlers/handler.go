package authhandlers

import (
	"Broker_backend/services/integration/partnerapi/internal/clients/authclient"
	"Broker_backend/services/integration/partnerapi/internal/transport/dto/authdto"
	"Broker_backend/services/integration/partnerapi/internal/transport/grpc/grpcerr"
	"Broker_backend/services/integration/partnerapi/internal/transport/http/httperr"
	"Broker_backend/services/integration/partnerapi/internal/transport/middleware"
	"errors"

	authv1 "Broker_backend/gen/auth/v1"

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
	req, ok := middleware.ValidatedBody[authdto.RegisterRequest](ctx)
	if !ok {
		return httperr.WriteBadRequest(ctx, "validated request is missing")
	}

	resp, err := h.authClient.Register(ctx.UserContext(), httpToGRPCRegister(req))
	if err != nil {
		h.logger.Warn("register grpc failed", zap.Error(err))
		middleware.AuditLog(ctx, h.logger, "auth.register.failed", zap.String("email", req.Email))
		return grpcerr.WriteGRPCError(ctx, err)
	}

	middleware.AuditLog(ctx, h.logger, "auth.register.succeeded", zap.String("email", req.Email))

	return ctx.Status(fiber.StatusCreated).JSON(grpcToHTTPTokenPair(resp))
}

func (h *AuthHandler) Login(ctx *fiber.Ctx) error {
	req, ok := middleware.ValidatedBody[authdto.LoginRequest](ctx)
	if !ok {
		return httperr.WriteBadRequest(ctx, "validated request is missing")
	}

	grpcReq := &authv1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
		DeviceId: req.DeviceID,
	}

	resp, err := h.authClient.Login(ctx.UserContext(), grpcReq)
	if err != nil {
		h.logger.Warn("login grpc failed", zap.Error(err))
		middleware.AuditLog(ctx, h.logger, "auth.login.failed", zap.String("email", req.Email))
		return grpcerr.WriteGRPCError(ctx, err)
	}

	middleware.AuditLog(ctx, h.logger, "auth.login.succeeded", zap.String("email", req.Email))

	return ctx.Status(fiber.StatusOK).JSON(grpcToHTTPTokenPair(resp))
}

func (h *AuthHandler) Refresh(ctx *fiber.Ctx) error {
	req, ok := middleware.ValidatedBody[authdto.RefreshRequest](ctx)
	if !ok {
		return httperr.WriteBadRequest(ctx, "validated request is missing")
	}

	grpcReq := &authv1.RefreshRequest{
		RefreshToken: req.RefreshToken,
		DeviceId:     req.DeviceID,
	}

	resp, err := h.authClient.Refresh(ctx.UserContext(), grpcReq)
	if err != nil {
		h.logger.Warn("refresh grpc failed", zap.Error(err))
		middleware.AuditLog(ctx, h.logger, "auth.refresh.failed", zap.String("device_id", req.DeviceID))
		return grpcerr.WriteGRPCError(ctx, err)
	}

	middleware.AuditLog(ctx, h.logger, "auth.refresh.succeeded", zap.String("device_id", req.DeviceID))

	return ctx.Status(fiber.StatusOK).JSON(grpcToHTTPTokenPair(resp))
}

func (h *AuthHandler) Logout(ctx *fiber.Ctx) error {
	req, ok := middleware.ValidatedBody[authdto.LogoutRequest](ctx)
	if !ok {
		return httperr.WriteBadRequest(ctx, "validated request is missing")
	}

	grpcReq := &authv1.LogoutRequest{
		RefreshToken: req.RefreshToken,
	}

	if _, err := h.authClient.Logout(ctx.UserContext(), grpcReq); err != nil {
		h.logger.Warn("logout grpc failed", zap.Error(err))
		middleware.AuditLog(ctx, h.logger, "auth.logout.failed", zap.String("device_id", req.DeviceID))
		return grpcerr.WriteGRPCError(ctx, err)
	}

	middleware.AuditLog(ctx, h.logger, "auth.logout.succeeded", zap.String("device_id", req.DeviceID))

	// proto LogoutResponse теперь пустой, поэтому device_id берём из запроса, all_device отдаём false.
	return ctx.Status(fiber.StatusOK).JSON(authdto.LogoutResponse{
		DeviceID: req.DeviceID,
	})
}

func (h *AuthHandler) Me(ctx *fiber.Ctx) error {
	claims, ok := middleware.CurrentClaims(ctx)
	if !ok {
		return httperr.WriteUnauthorized(ctx, "auth context is missing")
	}

	return ctx.Status(fiber.StatusOK).JSON(authdto.MeResponse{
		AgencyID: claims.AgencyID,
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

	middleware.AuditLog(ctx, h.logger, "admin.ping")

	return ctx.Status(fiber.StatusOK).JSON(authdto.AdminPingResponse{
		OK:       true,
		AgencyID: claims.AgencyID,
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

// tokenPair объединяет Register/Login/RefreshResponse — у всех есть только access/refresh токены.
// expires_in_sec из proto убрали, поэтому в DTO оно теперь всегда 0.
type tokenPair interface {
	GetAccessToken() string
	GetRefreshToken() string
}

func grpcToHTTPTokenPair(resp tokenPair) authdto.TokenPairResponse {
	return authdto.TokenPairResponse{
		Access:  resp.GetAccessToken(),
		Refresh: resp.GetRefreshToken(),
	}
}
