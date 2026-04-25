package auth

import (
	"context"

	"flowforge-automation-backend/pkg/model/domain"
	"github.com/google/uuid"
)

// Repository defines auth repository interface
type Repository interface {
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByIDAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*domain.User, error)
	CreateTenant(ctx context.Context, tenant *domain.Tenant) error
	CreateUser(ctx context.Context, user *domain.User) error
	GetTenantBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
}
