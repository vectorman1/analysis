package server

import (
	"context"

	"github.com/vectorman1/analysis/analysis-api/generated/user_service"
	"github.com/vectorman1/analysis/analysis-api/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServiceServer struct {
	userService *service.UserService
	user_service.UnimplementedUserServiceServer
}

func (s *UserServiceServer) Login(context.Context, *user_service.LoginRequest) (*user_service.LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (s *UserServiceServer) Register(context.Context, *user_service.RegisterRequest) (*user_service.RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
