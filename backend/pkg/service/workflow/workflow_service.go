package workflow

import (
	"context"

	"flowforge-automation-backend/pkg/model/domain"
	"flowforge-automation-backend/pkg/model/dto"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, tenantID, userID uuid.UUID, req *dto.CreateWorkflowRequest) (*domain.Workflow, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Workflow, int64, error)
	GetByID(ctx context.Context, tenantID, workflowID uuid.UUID) (*domain.Workflow, error)
	Update(ctx context.Context, tenantID, userID, workflowID uuid.UUID, req *dto.UpdateWorkflowRequest) (*domain.Workflow, error)
	Delete(ctx context.Context, tenantID, workflowID uuid.UUID) error
	ListVersions(ctx context.Context, tenantID, workflowID uuid.UUID, limit, offset int) ([]domain.WorkflowVersion, int64, error)
	Rollback(ctx context.Context, tenantID, userID, workflowID uuid.UUID, version int) (*domain.Workflow, error)
}
