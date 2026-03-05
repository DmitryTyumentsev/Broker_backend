package handlers

import (
	auth_grpc "Donate_backend/services/api-gateway/internal/clients/auth-service/auth-grpc"
	"Donate_backend/shared/pkg/config"
)

type Handlers struct {
	Config *config.Loader
	log    *zap.Logger
	Auth   *auth_grpc.Handlers
}
