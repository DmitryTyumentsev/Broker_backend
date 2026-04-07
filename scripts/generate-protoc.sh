protoc \
-I ../shared/pkg/grpc/proto \
--go_out=../shared/pkg/grpc/gen/auth/v1 \
--go-grpc_out=../shared/pkg/grpc/gen/auth/v1 \
../shared/pkg/grpc/proto/auth/v1/authv1.proto