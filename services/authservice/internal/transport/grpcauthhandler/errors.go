package grpcauthhandler

import (
	"errors"

	"Broker_backend/services/authservice/internal/domain"
	"Broker_backend/services/authservice/internal/usecases"

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
	case errors.Is(err, usecases.ErrEmailRequired):
		return codes.InvalidArgument, usecases.ErrEmailRequired.Error()
	case errors.Is(err, usecases.ErrEmailInvalid):
		return codes.InvalidArgument, usecases.ErrEmailInvalid.Error()
	case errors.Is(err, usecases.ErrPasswordRequired):
		return codes.InvalidArgument, usecases.ErrPasswordRequired.Error()
	case errors.Is(err, usecases.ErrPasswordWeak):
		return codes.InvalidArgument, usecases.ErrPasswordWeak.Error()
	case errors.Is(err, usecases.ErrDeviceIDRequired):
		return codes.InvalidArgument, usecases.ErrDeviceIDRequired.Error()
	case errors.Is(err, usecases.ErrFirstNameRequired):
		return codes.InvalidArgument, usecases.ErrFirstNameRequired.Error()
	case errors.Is(err, usecases.ErrLastNameRequired):
		return codes.InvalidArgument, usecases.ErrLastNameRequired.Error()
	case errors.Is(err, usecases.ErrRefreshRequired):
		return codes.InvalidArgument, usecases.ErrRefreshRequired.Error()
	case errors.Is(err, domain.ErrBadRequest):
		return codes.InvalidArgument, domain.ErrBadRequest.Error()
	case errors.Is(err, domain.ErrMustBeNotNull):
		return codes.InvalidArgument, domain.ErrMustBeNotNull.Error()
	case errors.Is(err, domain.ErrDeviceMismatch):
		return codes.InvalidArgument, domain.ErrDeviceMismatch.Error()

	case errors.Is(err, domain.ErrNotUniqueEmail), errors.Is(err, domain.ErrEmailExist):
		return codes.AlreadyExists, "email already exists"
	case errors.Is(err, domain.ErrNotUnique):
		return codes.AlreadyExists, "resource already exists"

	case errors.Is(err, domain.ErrUnauthenticated),
		errors.Is(err, domain.ErrPasswordWrong),
		errors.Is(err, domain.ErrSessionRevoked),
		errors.Is(err, domain.ErrSessionExpired):
		return codes.Unauthenticated, "invalid credentials or token"

	case errors.Is(err, domain.ErrNotFound):
		return codes.NotFound, domain.ErrNotFound.Error()

	case errors.Is(err, domain.ErrForbidden):
		return codes.PermissionDenied, domain.ErrForbidden.Error()

	default:
		return codes.Internal, "internal server error"
	}
}
