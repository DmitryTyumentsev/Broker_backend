protoc \
-I ./../shared/pkg/grpc/proto \
--go_out=./../tools/protoc-gen --go_opt=paths=source_relative \
--go_grpc_out=./../tools/protoc-gen --go_opt=paths=source_relative \
./../shared/pkg/grpc/proto/auth/v1/auth.proto