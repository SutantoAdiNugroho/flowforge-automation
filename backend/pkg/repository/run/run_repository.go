package run

import (
	"context"

	"flowforge-automation-backend/pkg/model/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, run *domain.WorkflowRun) error
	GetByID(ctx context.Context, runID uuid.UUID) (*domain.WorkflowRun, error)
	GetByIDAndTenant(ctx context.Context, runID, tenantID uuid.UUID) (*domain.WorkflowRun, error)
	ListByWorkflow(ctx context.Context, workflowID, tenantID uuid.UUID, limit, offset int) ([]domain.WorkflowRun, int64, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.WorkflowRun, int64, error)
	UpdateStatus(ctx context.Context, runID uuid.UUID, status string, errorMsg string) error
	CreateStepExecution(ctx context.Context, step *domain.StepExecution) error
	UpdateStepExecution(ctx context.Context, step *domain.StepExecution) error
	GetStepsByRunID(ctx context.Context, runID uuid.UUID) ([]domain.StepExecution, error)
	GetRunStats(ctx context.Context, tenantID uuid.UUID, hours int) (*RunStats, error)
}

type RunStats struct {
	TotalRuns    int64   `json:"total_runs"`
	SuccessRuns  int64   `json:"success_runs"`
	FailedRuns   int64   `json:"failed_runs"`
	SuccessRate  float64 `json:"success_rate"`
	AvgDurationMs int64  `json:"avg_duration_ms"`
}
