package grpchandler

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

	switch {
	case errors.Is(err, usecases.ErrEmailRequired),
		errors.Is(err, usecases.ErrEmailInvalid),
		errors.Is(err, usecases.ErrPasswordRequired),
		errors.Is(err, usecases.ErrPasswordWeak),
		errors.Is(err, usecases.ErrDeviceIDRequired),
		errors.Is(err, usecases.ErrFirstNameRequired),
		errors.Is(err, usecases.ErrLastNameRequired),
		errors.Is(err, usecases.ErrRefreshRequired),
		errors.Is(err, domain.ErrBadRequest),
		errors.Is(err, domain.ErrMustBeNotNull),
		errors.Is(err, domain.ErrDeviceMismatch):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, domain.ErrNotUniqueEmail),
		errors.Is(err, domain.ErrEmailExist),
		errors.Is(err, domain.ErrNotUnique):
		return status.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, domain.ErrUnauthenticated),
		errors.Is(err, domain.ErrPasswordWrong),
		errors.Is(err, domain.ErrSessionRevoked),
		errors.Is(err, domain.ErrSessionExpired):
		return status.Error(codes.Unauthenticated, "invalid credentials or token")

	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, domain.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())

	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
