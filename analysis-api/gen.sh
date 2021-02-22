go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc

mkdir -p api/swagger/v1

mkdir -p generated

mkdir -p generated/proto_models
protoc -I . --proto_path=api/proto/v1 \
            --proto_path=third_party \
            --go_out=generated/proto_models \
            --go_opt=paths=source_relative \
            models.proto


mkdir -p generated/symbol_service
# Symbols Service
protoc -I . --proto_path=api/proto/v1 \
            --proto_path=third_party \
            --go_out=generated/symbol_service \
            --go_opt=paths=source_relative \
            --go-grpc_out=generated/symbol_service \
            --go-grpc_opt=paths=source_relative \
            symbol_service.proto

protoc -I . --proto_path=api/proto/v1 \
            --proto_path=third_party \
            --openapiv2_out ./api/swagger/v1 \
            --openapiv2_opt logtostderr=true \
            symbol_service.proto

protoc -I . --proto_path=api/proto/v1 \
            --proto_path=third_party \
            --grpc-gateway_out ./generated/symbol_service \
            --grpc-gateway_opt paths=source_relative \
            --grpc-gateway_opt generate_unbound_methods=true \
            symbol_service.proto

mkdir -p generated/user_service
protoc -I . --proto_path=api/proto/v1 \
            --proto_path=third_party \
            --go_out=generated/user_service \
            --go_opt=paths=source_relative \
            --go-grpc_out=generated/user_service \
            --go-grpc_opt=paths=source_relative \
            user_service.proto

protoc -I . --proto_path=api/proto/v1 \
            --proto_path=third_party \
            --openapiv2_out ./api/swagger/v1 \
            --openapiv2_opt logtostderr=true \
             user_service.proto

protoc -I . --proto_path=api/proto/v1 \
            --proto_path=third_party \
            --grpc-gateway_out ./generated/user_service \
            --grpc-gateway_opt paths=source_relative \
            --grpc-gateway_opt generate_unbound_methods=true \
             user_service.proto

# Worker services

mkdir -p generated/trading212_service
protoc -I . --proto_path=api/proto/v1 \
            --proto_path=third_party \
            --go_out=generated/trading212_service \
            --go_opt=paths=source_relative \
            --go-grpc_out=generated/trading212_service \
            --go-grpc_opt=paths=source_relative \
            trading212_service.proto