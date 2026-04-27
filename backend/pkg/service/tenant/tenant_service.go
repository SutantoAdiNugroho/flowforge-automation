package tenant

import (
	"context"

	"flowforge-automation-backend/pkg/model/domain"
	"flowforge-automation-backend/pkg/model/dto"
	"flowforge-automation-backend/pkg/repository/tenant"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, req *dto.CreateTenantRequest) (*domain.Tenant, *domain.User, error)
	List(ctx context.Context, limit, offset int) ([]tenant.TenantWithStats, int64, error)
	GetByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error)
	Update(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateTenantRequest) (*domain.Tenant, error)
	Delete(ctx context.Context, tenantID uuid.UUID) error
}
