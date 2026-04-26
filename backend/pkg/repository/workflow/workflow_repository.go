package workflow

import (
	"context"

	"flowforge-automation-backend/pkg/model/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, workflow *domain.Workflow) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Workflow, int64, error)
	GetByIDAndTenant(ctx context.Context, workflowID, tenantID uuid.UUID) (*domain.Workflow, error)
	Update(ctx context.Context, workflow *domain.Workflow, tenantID uuid.UUID) error
	Delete(ctx context.Context, workflowID, tenantID uuid.UUID) error
}
