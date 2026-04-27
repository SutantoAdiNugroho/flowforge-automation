package dto

import "github.com/google/uuid"

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt int64     `json:"expires_at"`
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
}

// RegisterRequest represents register request payload
type RegisterRequest struct {
	TenantName string `json:"tenant_name" binding:"required,min=3"`
	TenantSlug string `json:"tenant_slug" binding:"required,min=3,alphanum"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
}

// RegisterResponse represents register response
type RegisterResponse struct {
	Token     string    `json:"token"`
	ExpiresAt int64     `json:"expires_at"`
	UserID    uuid.UUID `json:"user_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
