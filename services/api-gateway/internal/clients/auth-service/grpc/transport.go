package grpc

import authv1 "Donate_backend/shared/pkg/grpc/gen/auth/v1"

type Handlers struct{
	auth *authv1.AuthServiceClient
}

func NewHandler(auth *authv1.AuthServiceClient) *Handlers {
	return &Handlers{auth: auth,}
}

func NewAuthServiceClient() *authv1.AuthServiceClient {
	return &authv1.NewAuthServiceClient()//TODO: как тут создать экземляр *authv1.AuthServiceClient? я думал надо вызвать у них New функцию
}

func(h *Handlers) Register(req *authv1.RegisterRequest) (resp authv1.AuthPair, err error) {
	h.auth.
}