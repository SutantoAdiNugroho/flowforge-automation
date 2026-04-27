package user

import (
	"context"

	"flowforge-automation-backend/pkg/model/domain"
	"flowforge-automation-backend/pkg/model/dto"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, tenantID uuid.UUID, req *dto.CreateUserRequest) (*domain.User, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.User, int64, error)
	GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, tenantID, userID uuid.UUID, req *dto.UpdateUserRequest) (*domain.User, error)
	Delete(ctx context.Context, tenantID, userID uuid.UUID) error
}
