package tenant

import (
	"context"

	"flowforge-automation-backend/pkg/model/domain"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, tenant *domain.Tenant) error
	ListWithStats(ctx context.Context, limit, offset int) ([]TenantWithStats, int64, error)
	GetByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	Update(ctx context.Context, tenant *domain.Tenant) error
	Delete(ctx context.Context, tenantID uuid.UUID) error
}

type TenantWithStats struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	UserCount int64     `json:"user_count"`
	RunCount  int64     `json:"run_count"`
	CreatedAt string    `json:"created_at"`
}
