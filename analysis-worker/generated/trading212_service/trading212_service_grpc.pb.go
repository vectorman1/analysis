// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package trading212_service

import (
	context "context"
	proto_models "github.com/vectorman1/analysis/analysis-worker/generated/proto_models"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// Trading212ServiceClient is the client API for Trading212Service service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type Trading212ServiceClient interface {
	GetUpdatedSymbols(ctx context.Context, in *proto_models.Symbols, opts ...grpc.CallOption) (Trading212Service_GetUpdatedSymbolsClient, error)
	GetSymbols(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (Trading212Service_GetSymbolsClient, error)
}

type trading212ServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTrading212ServiceClient(cc grpc.ClientConnInterface) Trading212ServiceClient {
	return &trading212ServiceClient{cc}
}

func (c *trading212ServiceClient) GetUpdatedSymbols(ctx context.Context, in *proto_models.Symbols, opts ...grpc.CallOption) (Trading212Service_GetUpdatedSymbolsClient, error) {
	stream, err := c.cc.NewStream(ctx, &Trading212Service_ServiceDesc.Streams[0], "/v1.Trading212Service/GetUpdatedSymbols", opts...)
	if err != nil {
		return nil, err
	}
	x := &trading212ServiceGetUpdatedSymbolsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Trading212Service_GetUpdatedSymbolsClient interface {
	Recv() (*proto_models.Symbol, error)
	grpc.ClientStream
}

type trading212ServiceGetUpdatedSymbolsClient struct {
	grpc.ClientStream
}

func (x *trading212ServiceGetUpdatedSymbolsClient) Recv() (*proto_models.Symbol, error) {
	m := new(proto_models.Symbol)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *trading212ServiceClient) GetSymbols(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (Trading212Service_GetSymbolsClient, error) {
	stream, err := c.cc.NewStream(ctx, &Trading212Service_ServiceDesc.Streams[1], "/v1.Trading212Service/GetSymbols", opts...)
	if err != nil {
		return nil, err
	}
	x := &trading212ServiceGetSymbolsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Trading212Service_GetSymbolsClient interface {
	Recv() (*proto_models.Symbol, error)
	grpc.ClientStream
}

type trading212ServiceGetSymbolsClient struct {
	grpc.ClientStream
}

func (x *trading212ServiceGetSymbolsClient) Recv() (*proto_models.Symbol, error) {
	m := new(proto_models.Symbol)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Trading212ServiceServer is the server API for Trading212Service service.
// All implementations must embed UnimplementedTrading212ServiceServer
// for forward compatibility
type Trading212ServiceServer interface {
	GetUpdatedSymbols(*proto_models.Symbols, Trading212Service_GetUpdatedSymbolsServer) error
	GetSymbols(*GetRequest, Trading212Service_GetSymbolsServer) error
	mustEmbedUnimplementedTrading212ServiceServer()
}

// UnimplementedTrading212ServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTrading212ServiceServer struct {
}

func (UnimplementedTrading212ServiceServer) GetUpdatedSymbols(*proto_models.Symbols, Trading212Service_GetUpdatedSymbolsServer) error {
	return status.Errorf(codes.Unimplemented, "method GetUpdatedSymbols not implemented")
}
func (UnimplementedTrading212ServiceServer) GetSymbols(*GetRequest, Trading212Service_GetSymbolsServer) error {
	return status.Errorf(codes.Unimplemented, "method GetSymbols not implemented")
}
func (UnimplementedTrading212ServiceServer) mustEmbedUnimplementedTrading212ServiceServer() {}

// UnsafeTrading212ServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to Trading212ServiceServer will
// result in compilation errors.
type UnsafeTrading212ServiceServer interface {
	mustEmbedUnimplementedTrading212ServiceServer()
}

func RegisterTrading212ServiceServer(s grpc.ServiceRegistrar, srv Trading212ServiceServer) {
	s.RegisterService(&Trading212Service_ServiceDesc, srv)
}

func _Trading212Service_GetUpdatedSymbols_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(proto_models.Symbols)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(Trading212ServiceServer).GetUpdatedSymbols(m, &trading212ServiceGetUpdatedSymbolsServer{stream})
}

type Trading212Service_GetUpdatedSymbolsServer interface {
	Send(*proto_models.Symbol) error
	grpc.ServerStream
}

type trading212ServiceGetUpdatedSymbolsServer struct {
	grpc.ServerStream
}

func (x *trading212ServiceGetUpdatedSymbolsServer) Send(m *proto_models.Symbol) error {
	return x.ServerStream.SendMsg(m)
}

func _Trading212Service_GetSymbols_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(Trading212ServiceServer).GetSymbols(m, &trading212ServiceGetSymbolsServer{stream})
}

type Trading212Service_GetSymbolsServer interface {
	Send(*proto_models.Symbol) error
	grpc.ServerStream
}

type trading212ServiceGetSymbolsServer struct {
	grpc.ServerStream
}

func (x *trading212ServiceGetSymbolsServer) Send(m *proto_models.Symbol) error {
	return x.ServerStream.SendMsg(m)
}

// Trading212Service_ServiceDesc is the grpc.ServiceDesc for Trading212Service service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Trading212Service_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "v1.Trading212Service",
	HandlerType: (*Trading212ServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetUpdatedSymbols",
			Handler:       _Trading212Service_GetUpdatedSymbols_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetSymbols",
			Handler:       _Trading212Service_GetSymbols_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "trading212_service.proto",
}