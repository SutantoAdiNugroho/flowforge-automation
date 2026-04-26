package version

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

func NewVersionRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, v *domain.WorkflowVersion) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *repository) ListByWorkflow(ctx context.Context, workflowID uuid.UUID, limit, offset int) ([]domain.WorkflowVersion, int64, error) {
	var versions []domain.WorkflowVersion
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.WorkflowVersion{}).Where("workflow_id = ?", workflowID)
	q.Count(&total)

	err := q.Order("version DESC").Limit(limit).Offset(offset).Find(&versions).Error
	return versions, total, err
}

func (r *repository) GetByWorkflowAndVersion(ctx context.Context, workflowID uuid.UUID, ver int) (*domain.WorkflowVersion, error) {
	var v domain.WorkflowVersion
	err := r.db.WithContext(ctx).Where("workflow_id = ? AND version = ?", workflowID, ver).First(&v).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}
