package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"flowforge-automation-backend/internal/auth"
	"flowforge-automation-backend/pkg/model/domain"
	"flowforge-automation-backend/pkg/model/domain/enum"
	"flowforge-automation-backend/pkg/model/dto"
	authrepository "flowforge-automation-backend/pkg/repository/auth"
	"github.com/google/uuid"
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
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	// find user by email
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	// check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
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
	if req.Email == "" || req.Password == "" || req.TenantName == "" || req.TenantSlug == "" {
		return nil, errors.New("all fields are required")
	}

	existingTenant, err := s.repo.GetTenantBySlug(ctx, req.TenantSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to check tenant: %w", err)
	}
	if existingTenant != nil {
		return nil, errors.New("tenant slug already exists")
	}

	existingUser, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	tenantID := uuid.New()
	tenant := &domain.Tenant{
		BaseModel: domain.BaseModel{
			ID: tenantID,
		},
		Name: req.TenantName,
		Slug: req.TenantSlug,
	}

	if err := s.repo.CreateTenant(ctx, tenant); err != nil {
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
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         string(enum.UserRoleAdmin),
		IsActive:     true,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
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
