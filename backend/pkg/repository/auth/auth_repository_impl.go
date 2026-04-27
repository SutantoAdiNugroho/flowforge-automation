package auth

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

func NewAuthRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetUserByIDAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) CheckUserExists(ctx context.Context, userID, tenantID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ? AND tenant_id = ?", userID, tenantID).
		Count(&count).Error
	return count > 0, err
}

func (r *repository) CreateTenant(ctx context.Context, tenant *domain.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *repository) CreateUser(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *repository) GetTenantBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}
