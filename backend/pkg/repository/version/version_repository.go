package version

import (
	"context"

	"flowforge-automation-backend/pkg/model/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, version *domain.WorkflowVersion) error
	ListByWorkflow(ctx context.Context, workflowID uuid.UUID, limit, offset int) ([]domain.WorkflowVersion, int64, error)
	GetByWorkflowAndVersion(ctx context.Context, workflowID uuid.UUID, version int) (*domain.WorkflowVersion, error)
}
