protoc \
-I ../shared/pkg/grpc/proto \
--go_out=../shared/pkg/grpc/gen \
--go-grpc_out=../shared/pkg/grpc/gen \
../shared/pkg/grpc/proto/auth/v1/authv1.proto