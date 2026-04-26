package run

import (
	"context"
	"errors"
	"time"

	"flowforge-automation-backend/pkg/model/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRunRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, run *domain.WorkflowRun) error {
	return r.db.WithContext(ctx).Create(run).Error
}

func (r *repository) GetByID(ctx context.Context, runID uuid.UUID) (*domain.WorkflowRun, error) {
	var run domain.WorkflowRun
	err := r.db.WithContext(ctx).Where("id = ?", runID).First(&run).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &run, nil
}

func (r *repository) GetByIDAndTenant(ctx context.Context, runID, tenantID uuid.UUID) (*domain.WorkflowRun, error) {
	var run domain.WorkflowRun
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", runID, tenantID).First(&run).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &run, nil
}

func (r *repository) ListByWorkflow(ctx context.Context, workflowID, tenantID uuid.UUID, limit, offset int) ([]domain.WorkflowRun, int64, error) {
	var runs []domain.WorkflowRun
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.WorkflowRun{}).Where("workflow_id = ? AND tenant_id = ?", workflowID, tenantID)
	q.Count(&total)

	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&runs).Error
	return runs, total, err
}

func (r *repository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.WorkflowRun, int64, error) {
	var runs []domain.WorkflowRun
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.WorkflowRun{}).Where("tenant_id = ?", tenantID)
	q.Count(&total)

	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&runs).Error
	return runs, total, err
}

func (r *repository) UpdateStatus(ctx context.Context, runID uuid.UUID, status string, errorMsg string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}
	if status == "running" {
		now := time.Now()
		updates["started_at"] = now
	}
	if status == "success" || status == "failed" || status == "timeout" || status == "cancelled" {
		now := time.Now()
		updates["completed_at"] = now
	}
	return r.db.WithContext(ctx).Model(&domain.WorkflowRun{}).Where("id = ?", runID).Updates(updates).Error
}

func (r *repository) CreateStepExecution(ctx context.Context, step *domain.StepExecution) error {
	return r.db.WithContext(ctx).Create(step).Error
}

func (r *repository) UpdateStepExecution(ctx context.Context, step *domain.StepExecution) error {
	return r.db.WithContext(ctx).Model(&domain.StepExecution{}).Where("id = ?", step.ID).Updates(map[string]interface{}{
		"status":       step.Status,
		"completed_at": step.CompletedAt,
		"duration_ms":  step.DurationMs,
		"retry_count":  step.RetryCount,
		"output":       step.Output,
		"error":        step.Error,
		"updated_at":   time.Now(),
	}).Error
}

func (r *repository) GetStepsByRunID(ctx context.Context, runID uuid.UUID) ([]domain.StepExecution, error) {
	var steps []domain.StepExecution
	err := r.db.WithContext(ctx).Where("run_id = ?", runID).Order("created_at ASC").Find(&steps).Error
	return steps, err
}

func (r *repository) GetRunStats(ctx context.Context, tenantID uuid.UUID, hours int) (*RunStats, error) {
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	var stats RunStats

	r.db.WithContext(ctx).Model(&domain.WorkflowRun{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, since).
		Count(&stats.TotalRuns)

	r.db.WithContext(ctx).Model(&domain.WorkflowRun{}).
		Where("tenant_id = ? AND created_at >= ? AND status = 'success'", tenantID, since).
		Count(&stats.SuccessRuns)

	r.db.WithContext(ctx).Model(&domain.WorkflowRun{}).
		Where("tenant_id = ? AND created_at >= ? AND status = 'failed'", tenantID, since).
		Count(&stats.FailedRuns)

	if stats.TotalRuns > 0 {
		stats.SuccessRate = float64(stats.SuccessRuns) / float64(stats.TotalRuns) * 100
	}

	// avg duration of completed runs
	var avgResult struct{ Avg *float64 }
	r.db.WithContext(ctx).Model(&domain.WorkflowRun{}).
		Select("AVG(EXTRACT(EPOCH FROM (completed_at - started_at)) * 1000) as avg").
		Where("tenant_id = ? AND created_at >= ? AND completed_at IS NOT NULL AND started_at IS NOT NULL", tenantID, since).
		Scan(&avgResult)
	if avgResult.Avg != nil {
		stats.AvgDurationMs = int64(*avgResult.Avg)
	}

	return &stats, nil
}
