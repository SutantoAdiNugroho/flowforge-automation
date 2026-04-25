package auth

import (
	"net/http"

	"flowforge-automation-backend/pkg/model/domain/enum"
	"flowforge-automation-backend/pkg/model/dto"
	authservice "flowforge-automation-backend/pkg/service/auth"
	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	authService authservice.Service
}

func NewAuthController(authService authservice.Service) *Controller {
	return &Controller{
		authService: authService,
	}
}

// Login handles user login
// @Summary Login
// @Description User login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/auth/login [post]
func (c *Controller) Login(ctx *fiber.Ctx) error {
	req := &dto.LoginRequest{}
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   string(enum.AuthErrorInvalidRequest),
			Message: "Invalid request body",
		})
	}

	resp, err := c.authService.Login(ctx.Context(), req)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error:   string(enum.AuthErrorLoginFailed),
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(resp)
}

// Register handles user registration
// @Summary Register
// @Description Create new tenant and user account
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dto.RegisterRequest true "Registration data"
// @Success 201 {object} dto.RegisterResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/auth/register [post]
func (c *Controller) Register(ctx *fiber.Ctx) error {
	req := &dto.RegisterRequest{}
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{
			Error:   string(enum.AuthErrorInvalidRequest),
			Message: "Invalid request body",
		})
	}

	resp, err := c.authService.Register(ctx.Context(), req)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "email already registered" || err.Error() == "tenant slug already exists" {
			statusCode = http.StatusConflict
		}
		return ctx.Status(statusCode).JSON(dto.ErrorResponse{
			Error:   string(enum.AuthErrorRegisterFailed),
			Message: err.Error(),
		})
	}

	return ctx.Status(http.StatusCreated).JSON(resp)
}
