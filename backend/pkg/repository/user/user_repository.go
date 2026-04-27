package user

import (
	"context"

	"flowforge-automation-backend/pkg/model/domain"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, user *domain.User) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.User, int64, error)
	GetByIDAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByEmailAndTenant(ctx context.Context, email string, tenantID uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, user *domain.User, tenantID uuid.UUID) error
	Delete(ctx context.Context, userID, tenantID uuid.UUID) error
}
