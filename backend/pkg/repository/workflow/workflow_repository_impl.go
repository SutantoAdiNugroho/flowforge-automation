package workflow

import (
	"context"
	"errors"

	"flowforge-automation-backend/pkg/model/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewWorkflowRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, workflow *domain.Workflow) error {
	return r.db.WithContext(ctx).Create(workflow).Error
}

func (r *repository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Workflow, int64, error) {
	var workflows []domain.Workflow
	var total int64
	
	q := r.db.WithContext(ctx).Model(&domain.Workflow{}).Where("tenant_id = ?", tenantID)
	q.Count(&total)

	err := q.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&workflows).Error
	if err != nil {
		return nil, 0, err
	}
	return workflows, total, nil
}

func (r *repository) GetByIDAndTenant(ctx context.Context, workflowID, tenantID uuid.UUID) (*domain.Workflow, error) {
	var workflow domain.Workflow
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", workflowID, tenantID).
		First(&workflow).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &workflow, nil
}

func (r *repository) Update(ctx context.Context, workflow *domain.Workflow, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.Workflow{}).
		Where("id = ? AND tenant_id = ?", workflow.ID, tenantID).
		Updates(workflow).Error
}

func (r *repository) Delete(ctx context.Context, workflowID, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", workflowID, tenantID).
		Delete(&domain.Workflow{}).Error
}
