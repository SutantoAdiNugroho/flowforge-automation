package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"flowforge-automation-backend/pkg/model/domain/enum"
	"flowforge-automation-backend/pkg/model/dto"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ContextKey string

const (
	UserIDKey     ContextKey = "user_id"
	TenantIDKey   ContextKey = "tenant_id"
	EmailKey      ContextKey = "email"
	RoleKey       ContextKey = "role"
	UserClaimsKey ContextKey = "user_claims"
)

type UserLookup interface {
	CheckUserExists(ctx context.Context, userID, tenantID uuid.UUID) (bool, error)
}

func Middleware(jwtManager *JWTManager, repo UserLookup) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   string(enum.AuthErrorMissingAuth),
				Message: "Authorization header missing",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   string(enum.AuthErrorInvalidAuthFormat),
				Message: "Invalid authorization format",
			})
		}

		tokenString := parts[1]

		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   string(enum.AuthErrorInvalidToken),
				Message: "Invalid or expired token",
			})
		}

		if repo != nil {
			exists, err := repo.CheckUserExists(ctx.Context(), claims.UserID, claims.TenantID)
			if err != nil || !exists {
				return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{
					Error:   string(enum.AuthErrorUnauthorized),
					Message: "User or tenant no longer exists. Please login again.",
				})
			}
		}

		ctx.Locals(string(UserIDKey), claims.UserID.String())
		ctx.Locals(string(TenantIDKey), claims.TenantID.String())
		ctx.Locals(string(EmailKey), claims.Email)
		ctx.Locals(string(RoleKey), claims.Role)
		ctx.Locals(string(UserClaimsKey), claims)

		return ctx.Next()
	}
}

// RoleBasedMiddleware returns role-based access control middleware
func RoleBasedMiddleware(allowedRoles ...string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		role := ctx.Locals(string(RoleKey))
		if role == nil {
			return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{
				Error:   string(enum.AuthErrorUnauthorized),
				Message: "User not authenticated",
			})
		}

		userRole := role.(string)
		
		if userRole == string(enum.UserRoleSuperAdmin) {
			return ctx.Next()
		}

		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				return ctx.Next()
			}
		}

		return ctx.Status(http.StatusForbidden).JSON(dto.ErrorResponse{
			Error:   string(enum.AuthErrorForbidden),
			Message: "Insufficient permissions",
		})
	}
}

func GetUserID(ctx *fiber.Ctx) (string, error) {
	userID := ctx.Locals(string(UserIDKey))
	if userID == nil {
		return "", errors.New("user id not found in context")
	}
	return userID.(string), nil
}

func GetTenantID(ctx *fiber.Ctx) (string, error) {
	tenantID := ctx.Locals(string(TenantIDKey))
	if tenantID == nil {
		return "", errors.New("tenant id not found in context")
	}
	return tenantID.(string), nil
}

func GetRole(ctx *fiber.Ctx) (string, error) {
	role := ctx.Locals(string(RoleKey))
	if role == nil {
		return "", errors.New("role not found in context")
	}
	return role.(string), nil
}
