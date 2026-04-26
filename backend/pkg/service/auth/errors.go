package auth

import "errors"

var (
	ErrInvalidLoginRequest    = errors.New("invalid login request")
	ErrInvalidRegisterRequest = errors.New("invalid register request")
	ErrInvalidCredentials     = errors.New("invalid email or password")
	ErrInactiveUser           = errors.New("user account is inactive")
	ErrTenantSlugExists       = errors.New("tenant slug already exists")
	ErrEmailAlreadyRegistered = errors.New("email already registered")
)
