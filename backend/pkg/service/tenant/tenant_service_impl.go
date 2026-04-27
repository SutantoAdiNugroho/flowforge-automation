package tenant

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"flowforge-automation-backend/pkg/model/domain"
	"flowforge-automation-backend/pkg/model/domain/enum"
	"flowforge-automation-backend/pkg/model/dto"
	"flowforge-automation-backend/pkg/repository/tenant"
	userrepo "flowforge-automation-backend/pkg/repository/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrTenantNotFound       = errors.New("tenant not found")
	ErrInvalidCreateRequest = errors.New("invalid create request")
	ErrTenantSlugExists     = errors.New("tenant slug already exists")
	ErrInvalidUpdateRequest = errors.New("invalid update request")
	ErrInvalidTenantSlug    = errors.New("invalid tenant slug format")
)

type service struct {
	tenantRepo tenant.Repository
	userRepo   userrepo.Repository
}

func NewTenantService(tenantRepo tenant.Repository, userRepo userrepo.Repository) Service {
	return &service{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func isValidTenantSlug(slug string) bool {
	if len(slug) < 3 || len(slug) > 50 {
		return false
	}
	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

func (s *service) Create(ctx context.Context, req *dto.CreateTenantRequest) (*domain.Tenant, *domain.User, error) {
	if req == nil || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Slug) == "" ||
		strings.TrimSpace(req.AdminEmail) == "" || strings.TrimSpace(req.AdminPassword) == "" {
		return nil, nil, ErrInvalidCreateRequest
	}

	slug := strings.ToLower(strings.TrimSpace(req.Slug))
	if !isValidTenantSlug(slug) {
		return nil, nil, ErrInvalidTenantSlug
	}

	if len(req.AdminPassword) < 6 {
		return nil, nil, errors.New("password must be at least 6 characters")
	}

	existing, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check tenant: %w", err)
	}
	if existing != nil {
		return nil, nil, ErrTenantSlugExists
	}

	tenantID := uuid.New()
	tenantObj := &domain.Tenant{
		BaseModel: domain.BaseModel{ID: tenantID},
		Name:      strings.TrimSpace(req.Name),
		Slug:      slug,
	}

	if err := s.tenantRepo.Create(ctx, tenantObj); err != nil {
		if isUniqueViolation(err) {
			return nil, nil, ErrTenantSlugExists
		}
		return nil, nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	userID := uuid.New()
	adminUser := &domain.User{
		BaseModel:    domain.BaseModel{ID: userID},
		TenantID:     tenantID,
		Email:        strings.ToLower(strings.TrimSpace(req.AdminEmail)),
		PasswordHash: string(hashedPassword),
		Role:         string(enum.UserRoleAdmin),
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, adminUser); err != nil {
		if isUniqueViolation(err) {
			return tenantObj, nil, errors.New("admin email already registered in another tenant")
		}
		return tenantObj, nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	return tenantObj, adminUser, nil
}

func (s *service) List(ctx context.Context, limit, offset int) ([]tenant.TenantWithStats, int64, error) {
	return s.tenantRepo.ListWithStats(ctx, limit, offset)
}

func (s *service) GetByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	tenantObj, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if tenantObj == nil {
		return nil, ErrTenantNotFound
	}
	return tenantObj, nil
}

func (s *service) Update(ctx context.Context, tenantID uuid.UUID, req *dto.UpdateTenantRequest) (*domain.Tenant, error) {
	if req == nil || strings.TrimSpace(req.Name) == "" {
		return nil, ErrInvalidUpdateRequest
	}

	existing, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrTenantNotFound
	}

	existing.Name = strings.TrimSpace(req.Name)
	if err := s.tenantRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return existing, nil
}

func (s *service) Delete(ctx context.Context, tenantID uuid.UUID) error {
	existing, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrTenantNotFound
	}

	return s.tenantRepo.Delete(ctx, tenantID)
}
