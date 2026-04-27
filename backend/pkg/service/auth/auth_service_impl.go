package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"flowforge-automation-backend/internal/auth"
	"flowforge-automation-backend/pkg/model/domain"
	"flowforge-automation-backend/pkg/model/domain/enum"
	"flowforge-automation-backend/pkg/model/dto"
	authrepository "flowforge-automation-backend/pkg/repository/auth"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	repo     authrepository.Repository
	jwtMgr   *auth.JWTManager
	tokenTTL time.Duration
}

func NewAuthService(repo authrepository.Repository, jwtSecretKey string, tokenTTL time.Duration) Service {
	return &service{
		repo:     repo,
		jwtMgr:   auth.NewJWTManager(jwtSecretKey),
		tokenTTL: tokenTTL,
	}
}

func (s *service) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	if req == nil {
		return nil, ErrInvalidLoginRequest
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || strings.TrimSpace(req.Password) == "" {
		return nil, ErrInvalidLoginRequest
	}

	if !isValidEmail(email) {
		return nil, ErrInvalidLoginRequest
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// check if user is active
	if !user.IsActive {
		return nil, ErrInactiveUser
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	expiresAt := time.Now().Add(s.tokenTTL)
	token, err := s.jwtMgr.GenerateToken(user.ID, user.TenantID, user.Email, user.Role, s.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
	}, nil
}

func (s *service) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	if req == nil {
		return nil, ErrInvalidRegisterRequest
	}

	tenantSlug := normalizeSlug(req.TenantSlug)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.TenantName) == "" || tenantSlug == "" {
		return nil, ErrInvalidRegisterRequest
	}

	if !isValidEmail(email) || len(req.Password) < 6 || !isValidTenantSlug(tenantSlug) {
		return nil, ErrInvalidRegisterRequest
	}

	existingUser, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyRegistered
	}

	existingTenant, err := s.repo.GetTenantBySlug(ctx, tenantSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to check tenant: %w", err)
	}
	if existingTenant != nil {
		return nil, ErrTenantSlugExists
	}

	tenantID := uuid.New()
	tenant := &domain.Tenant{
		BaseModel: domain.BaseModel{
			ID: tenantID,
		},
		Name: req.TenantName,
		Slug: tenantSlug,
	}

	if err := s.repo.CreateTenant(ctx, tenant); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrTenantSlugExists
		}
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	userID := uuid.New()
	user := &domain.User{
		BaseModel: domain.BaseModel{
			ID: userID,
		},
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         string(enum.UserRoleAdmin),
		IsActive:     true,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailAlreadyRegistered
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	expiresAt := time.Now().Add(s.tokenTTL)
	token, err := s.jwtMgr.GenerateToken(userID, tenantID, user.Email, user.Role, s.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.RegisterResponse{
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
		UserID:    userID,
		TenantID:  tenantID,
		Email:     user.Email,
		Role:      user.Role,
	}, nil
}

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
