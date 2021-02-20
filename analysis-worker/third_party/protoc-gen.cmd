protoc --proto_path=api/proto/v2 --proto_path=third_party --go_out=./generated --go-grpc_out=./generated symbols_service.proto
protoc --proto_path=api/proto/v2 --proto_path=third_party --grpc-gateway_out=logtostderr=true:generated symbols_service.proto
protoc --proto_path=api/proto/v2 --proto_path=third_party --swagger_out=logtostderr=true:api/swagger/v2 symbols_service.proto