package grpchandler

import "google.golang.org/grpc/status"

func MapError(err error) error {
	st := status.Convert(err)
	status.New()
}
