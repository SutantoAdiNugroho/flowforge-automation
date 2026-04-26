package auth

import (
	"context"

	"flowforge-automation-backend/pkg/model/dto"
)

type Service interface {
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error)
}
