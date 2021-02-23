go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc

mkdir -p generated

mkdir -p generated/proto_models
protoc -I . --proto_path=service/proto \
            --proto_path=third_party \
            --go_out=generated/proto_models \
            --go_opt=paths=source_relative \
            models.proto


mkdir -p generated/trading212_service
protoc -I . --proto_path=service/proto \
            --proto_path=third_party \
            --go_out=generated/trading212_service \
            --go_opt=paths=source_relative \
            --go-grpc_out=generated/trading212_service \
            --go-grpc_opt=paths=source_relative \
            trading212_service.proto