package run

import (
	"context"

	"flowforge-automation-backend/pkg/model/domain"
	"flowforge-automation-backend/pkg/model/dto"
	runrepo "flowforge-automation-backend/pkg/repository/run"
	"github.com/google/uuid"
)

type Service interface {
	TriggerRun(ctx context.Context, tenantID, userID, workflowID uuid.UUID, req *dto.TriggerRunRequest) (*domain.WorkflowRun, error)
	GetRun(ctx context.Context, tenantID, runID uuid.UUID) (*dto.RunDetailResponse, error)
	ListRunsByWorkflow(ctx context.Context, tenantID, workflowID uuid.UUID, limit, offset int) ([]domain.WorkflowRun, int64, error)
	ListRunsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.WorkflowRun, int64, error)
	CancelRun(ctx context.Context, tenantID, runID uuid.UUID) error
	GetStats(ctx context.Context, tenantID uuid.UUID, hours int) (*runrepo.RunStats, error)
}
