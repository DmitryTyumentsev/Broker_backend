package brokerhandler

import (
	"Broker_backend/services/integration/brokerservice/internal/domain"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapError(err error) error {
	if err == nil {
		return nil
	}

	code, message := grpcStatusFromError(err)
	return status.Error(code, message)
}

func grpcStatusFromError(err error) (codes.Code, string) {
	switch {
	//case errors.Is(err, usecases.ErrRefreshRequired):
	//	return codes.InvalidArgument, usecases.ErrRefreshRequired.Error()
	case errors.Is(err, domain.ErrBadRequest):
		return codes.InvalidArgument, domain.ErrBadRequest.Error()
	case errors.Is(err, domain.ErrMustBeNotNull):
		return codes.InvalidArgument, domain.ErrMustBeNotNull.Error()
	case errors.Is(err, domain.ErrNotUnique):
		return codes.AlreadyExists, "resource already exists"

	case errors.Is(err, domain.ErrUnauthenticated):
		return codes.Unauthenticated, "invalid credentials or token"

	case errors.Is(err, domain.ErrNotFound):
		return codes.NotFound, domain.ErrNotFound.Error()

	case errors.Is(err, domain.ErrForbidden):
		return codes.PermissionDenied, domain.ErrForbidden.Error()

	default:
		return codes.Internal, "internal server error"
	}
}
