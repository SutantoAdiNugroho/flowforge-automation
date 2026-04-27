package tenant

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

func NewTenantRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, tenant *domain.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *repository) ListWithStats(ctx context.Context, limit, offset int) ([]TenantWithStats, int64, error) {
	var tenants []TenantWithStats
	var total int64

	err := r.db.WithContext(ctx).
		Raw(`
			SELECT 
				t.id, 
				t.name, 
				t.slug,
				COALESCE(COUNT(DISTINCT u.id), 0) as user_count,
				COALESCE(COUNT(DISTINCT r.id), 0) as run_count,
				t.created_at
			FROM tenants t
			LEFT JOIN users u ON u.tenant_id = t.id
			LEFT JOIN workflow_runs r ON r.tenant_id = t.id
			GROUP BY t.id, t.name, t.slug, t.created_at
			ORDER BY t.created_at DESC
			LIMIT ? OFFSET ?
		`, limit, offset).
		Scan(&tenants).Error

	if err != nil {
		return nil, 0, err
	}

	var countResult []struct {
		Total int64
	}
	if err := r.db.WithContext(ctx).
		Raw(`
			SELECT COUNT(DISTINCT t.id) as total
			FROM tenants t
		`).
		Scan(&countResult).Error; err != nil {
		return tenants, 0, err
	}

	if len(countResult) > 0 {
		total = countResult[0].Total
	}

	return tenants, total, nil
}

func (r *repository) GetByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := r.db.WithContext(ctx).
		Where("id = ?", tenantID).
		First(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *repository) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := r.db.WithContext(ctx).
		Where("slug = ?", slug).
		First(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *repository) Update(ctx context.Context, tenant *domain.Tenant) error {
	return r.db.WithContext(ctx).
		Model(&domain.Tenant{}).
		Where("id = ?", tenant.ID).
		Updates(tenant).Error
}

func (r *repository) Delete(ctx context.Context, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", tenantID).
		Delete(&domain.Tenant{}).Error
}
