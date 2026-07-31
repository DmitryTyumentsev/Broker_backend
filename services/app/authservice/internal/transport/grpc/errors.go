package grpc

import (
	"Broker_backend/services/app/authservice/internal/domain"
	"Broker_backend/services/app/authservice/internal/usecase"
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
	case errors.Is(err, usecase.ErrEmailRequired):
		return codes.InvalidArgument, usecase.ErrEmailRequired.Error()
	case errors.Is(err, usecase.ErrEmailInvalid):
		return codes.InvalidArgument, usecase.ErrEmailInvalid.Error()
	case errors.Is(err, usecase.ErrPasswordRequired):
		return codes.InvalidArgument, usecase.ErrPasswordRequired.Error()
	case errors.Is(err, usecase.ErrPasswordWeak):
		return codes.InvalidArgument, usecase.ErrPasswordWeak.Error()
	case errors.Is(err, usecase.ErrDeviceIDRequired):
		return codes.InvalidArgument, usecase.ErrDeviceIDRequired.Error()
	case errors.Is(err, usecase.ErrFirstNameRequired):
		return codes.InvalidArgument, usecase.ErrFirstNameRequired.Error()
	case errors.Is(err, usecase.ErrLastNameRequired):
		return codes.InvalidArgument, usecase.ErrLastNameRequired.Error()
	case errors.Is(err, usecase.ErrRefreshRequired):
		return codes.InvalidArgument, usecase.ErrRefreshRequired.Error()
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
