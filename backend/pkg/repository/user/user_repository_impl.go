package user

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

func NewUserRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *repository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.User{}).Where("tenant_id = ?", tenantID)
	q.Count(&total)

	err := q.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *repository) GetByIDAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", userID, tenantID).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetByEmailAndTenant(ctx context.Context, email string, tenantID uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Where("email = ? AND tenant_id = ?", email, tenantID).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) Update(ctx context.Context, user *domain.User, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ? AND tenant_id = ?", user.ID, tenantID).
		Updates(user).Error
}

func (r *repository) Delete(ctx context.Context, userID, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", userID, tenantID).
		Delete(&domain.User{}).Error
}
