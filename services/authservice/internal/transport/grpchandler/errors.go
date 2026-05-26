package grpchandler

import (
	"errors"

	"Donate_backend/services/authservice/internal/domain"
	"Donate_backend/services/authservice/internal/usecases"

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
		errors.Is(err, domain.ErrBadRequest),
		errors.Is(err, domain.ErrMustBeNotNull):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, domain.ErrEmailExist),
		errors.Is(err, domain.ErrUsernameExist),
		errors.Is(err, domain.ErrNotUnique):
		return status.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, domain.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, err.Error())

	case errors.Is(err, domain.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())

	default:
		return status.Error(codes.Internal, err.Error())
	}
}
