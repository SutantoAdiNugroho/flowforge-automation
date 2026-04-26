package auth

import (
	"context"
	"testing"
	"time"

	"flowforge-automation-backend/pkg/model/domain"
	"flowforge-automation-backend/pkg/model/domain/enum"
	"flowforge-automation-backend/pkg/model/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type fakeAuthRepo struct {
	tenantBySlug map[string]*domain.Tenant
	userByKey    map[string]*domain.User
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{
		tenantBySlug: make(map[string]*domain.Tenant),
		userByKey:    make(map[string]*domain.User),
	}
}

func userKey(email string, tenantID uuid.UUID) string {
	return email + "|" + tenantID.String()
}

func (r *fakeAuthRepo) GetUserByEmailAndTenant(ctx context.Context, email string, tenantID uuid.UUID) (*domain.User, error) {
	return r.userByKey[userKey(email, tenantID)], nil
}

func (r *fakeAuthRepo) GetUserByIDAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*domain.User, error) {
	for _, u := range r.userByKey {
		if u.ID == userID && u.TenantID == tenantID {
			return u, nil
		}
	}
	return nil, nil
}

func (r *fakeAuthRepo) CreateTenant(ctx context.Context, tenant *domain.Tenant) error {
	r.tenantBySlug[tenant.Slug] = tenant
	return nil
}

func (r *fakeAuthRepo) CreateUser(ctx context.Context, user *domain.User) error {
	r.userByKey[userKey(user.Email, user.TenantID)] = user
	return nil
}

func (r *fakeAuthRepo) GetTenantBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	return r.tenantBySlug[slug], nil
}

func TestLoginSuccess(t *testing.T) {
	repo := newFakeAuthRepo()
	tenantID := uuid.New()
	repo.tenantBySlug["acme"] = &domain.Tenant{BaseModel: domain.BaseModel{ID: tenantID}, Name: "Acme", Slug: "acme"}

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	userID := uuid.New()
	repo.userByKey[userKey("admin@acme.com", tenantID)] = &domain.User{
		BaseModel: domain.BaseModel{ID: userID},
		TenantID:     tenantID,
		Email:        "admin@acme.com",
		PasswordHash: string(hash),
		Role:         string(enum.UserRoleAdmin),
		IsActive:     true,
	}

	svc := NewAuthService(repo, "test-secret", 24*time.Hour)

	res, err := svc.Login(context.Background(), &dto.LoginRequest{
		TenantSlug: "acme",
		Email:      "admin@acme.com",
		Password:   "password123",
	})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res.Token)
	assert.Equal(t, userID, res.UserID)
	assert.Equal(t, "admin@acme.com", res.Email)
}

func TestLoginInvalidCredentials(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewAuthService(repo, "test-secret", 24*time.Hour)

	_, err := svc.Login(context.Background(), &dto.LoginRequest{
		TenantSlug: "unknown",
		Email:      "admin@acme.com",
		Password:   "password123",
	})

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestRegisterCreatesTenantAndAdmin(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewAuthService(repo, "test-secret", 24*time.Hour)

	res, err := svc.Register(context.Background(), &dto.RegisterRequest{
		TenantName: "Acme",
		TenantSlug: "acme",
		Email:      "owner@acme.com",
		Password:   "password123",
	})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "owner@acme.com", res.Email)
	assert.Equal(t, string(enum.UserRoleAdmin), res.Role)
	assert.NotNil(t, repo.tenantBySlug["acme"])
}
