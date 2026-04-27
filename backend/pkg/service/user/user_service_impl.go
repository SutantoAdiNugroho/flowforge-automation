package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"flowforge-automation-backend/pkg/model/domain"
	"flowforge-automation-backend/pkg/model/domain/enum"
	"flowforge-automation-backend/pkg/model/dto"
	userrepo "flowforge-automation-backend/pkg/repository/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidRole          = errors.New("invalid role")
	ErrInvalidUpdateRequest = errors.New("invalid update request")
	ErrInvalidCreateRequest = errors.New("invalid create request")
	ErrEmailAlreadyExists   = errors.New("email already exists")
)

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return strings.Contains(strings.ToLower(err.Error()), "duplicate")
}

type service struct {
	repo userrepo.Repository
}

func NewUserService(repo userrepo.Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, tenantID uuid.UUID, req *dto.CreateUserRequest) (*domain.User, error) {
	if req == nil || strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.Role) == "" {
		return nil, ErrInvalidCreateRequest
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	role := strings.ToLower(strings.TrimSpace(req.Role))

	if len(req.Password) < 6 {
		return nil, ErrInvalidCreateRequest
	}

	if role != string(enum.UserRoleAdmin) && role != string(enum.UserRoleEditor) && role != string(enum.UserRoleViewer) {
		return nil, ErrInvalidRole
	}

	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		BaseModel:    domain.BaseModel{ID: uuid.New()},
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *service) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.User, int64, error) {
	return s.repo.ListByTenant(ctx, tenantID, limit, offset)
}

func (s *service) GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error) {
	user, err := s.repo.GetByIDAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *service) Update(ctx context.Context, tenantID, userID uuid.UUID, req *dto.UpdateUserRequest) (*domain.User, error) {
	if req == nil || strings.TrimSpace(req.Role) == "" {
		return nil, ErrInvalidUpdateRequest
	}

	role := strings.ToLower(strings.TrimSpace(req.Role))
	if role != string(enum.UserRoleAdmin) && role != string(enum.UserRoleEditor) && role != string(enum.UserRoleViewer) {
		return nil, ErrInvalidRole
	}

	existing, err := s.repo.GetByIDAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrUserNotFound
	}

	existing.Role = role

	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, existing, tenantID); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *service) Delete(ctx context.Context, tenantID, userID uuid.UUID) error {
	existing, err := s.repo.GetByIDAndTenant(ctx, userID, tenantID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrUserNotFound
	}

	return s.repo.Delete(ctx, userID, tenantID)
}
