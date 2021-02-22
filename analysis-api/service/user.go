package service

import (
	"github.com/vectorman1/analysis/analysis-api/db"
	"github.com/vectorman1/analysis/analysis-api/generated/user_service"
)

type userService interface {
	Login(*user_service.LoginRequest) (*user_service.LoginResponse, error)
	Register(*user_service.RegisterRequest) (*user_service.RegisterResponse, error)
}

type UserService struct {
	userService
	userRepository *db.UserRepository
}

// Login attempts to find a user with a matching username and hashed password
// and returns a response with a Token or an error
func (s *UserService) Login(request *user_service.LoginRequest) (*user_service.LoginResponse, error) {
	_, err := s.userRepository.Login(request.Username, request.Password)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Register attempts to create a User with the corresponding
// username and hashing the password.
func (s *UserService) Register(request *user_service.RegisterRequest) (*user_service.RegisterResponse, error) {
	_, err := s.userRepository.Register(request.Username, request.Password)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
