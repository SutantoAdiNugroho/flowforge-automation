package user

import (
	"errors"
	"net/http"
	"strconv"

	internalauth "flowforge-automation-backend/internal/auth"
	"flowforge-automation-backend/pkg/model/dto"
	"flowforge-automation-backend/pkg/model/dto/response"
	userservice "flowforge-automation-backend/pkg/service/user"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Controller struct {
	userService userservice.Service
}

func NewUserController(userService userservice.Service) *Controller {
	return &Controller{userService: userService}
}

func (c *Controller) Create(ctx *fiber.Ctx) error {
	tenantID, _, err := getAuthContextIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	req := &dto.CreateUserRequest{}
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid request body"})
	}

	user, err := c.userService.Create(ctx.Context(), tenantID, req)
	if err != nil {
		if errors.Is(err, userservice.ErrInvalidCreateRequest) {
			return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: err.Error()})
		}
		if errors.Is(err, userservice.ErrInvalidRole) {
			return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_ROLE", Message: err.Error()})
		}
		if errors.Is(err, userservice.ErrEmailAlreadyExists) {
			return ctx.Status(http.StatusConflict).JSON(dto.ErrorResponse{Error: "EMAIL_EXISTS", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "USER_CREATE_FAILED", Message: "failed to create user"})
	}

	return ctx.Status(http.StatusCreated).JSON(user)
}

func (c *Controller) List(ctx *fiber.Ctx) error {
	tenantID, _, err := getAuthContextIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size", "20"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	users, total, err := c.userService.ListByTenant(ctx.Context(), tenantID, pageSize, offset)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "USER_LIST_FAILED", Message: "failed to list users"})
	}

	return ctx.Status(http.StatusOK).JSON(response.NewPaginationResponse(page, pageSize, int(total), users))
}

func (c *Controller) GetByID(ctx *fiber.Ctx) error {
	tenantID, _, err := getAuthContextIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	userID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid user id"})
	}

	user, err := c.userService.GetByID(ctx.Context(), tenantID, userID)
	if err != nil {
		if errors.Is(err, userservice.ErrUserNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "USER_NOT_FOUND", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "USER_GET_FAILED", Message: "failed to get user"})
	}

	return ctx.Status(http.StatusOK).JSON(user)
}

func (c *Controller) Update(ctx *fiber.Ctx) error {
	tenantID, _, err := getAuthContextIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	userID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid user id"})
	}

	req := &dto.UpdateUserRequest{}
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid request body"})
	}

	user, err := c.userService.Update(ctx.Context(), tenantID, userID, req)
	if err != nil {
		if errors.Is(err, userservice.ErrUserNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "USER_NOT_FOUND", Message: err.Error()})
		}
		if errors.Is(err, userservice.ErrInvalidRole) {
			return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_ROLE", Message: err.Error()})
		}
		if errors.Is(err, userservice.ErrInvalidUpdateRequest) {
			return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "USER_UPDATE_FAILED", Message: "failed to update user"})
	}

	return ctx.Status(http.StatusOK).JSON(user)
}

func (c *Controller) Delete(ctx *fiber.Ctx) error {
	tenantID, _, err := getAuthContextIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	userID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid user id"})
	}

	err = c.userService.Delete(ctx.Context(), tenantID, userID)
	if err != nil {
		if errors.Is(err, userservice.ErrUserNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "USER_NOT_FOUND", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "USER_DELETE_FAILED", Message: "failed to delete user"})
	}

	return ctx.Status(http.StatusNoContent).Send(nil)
}

func getAuthContextIDs(ctx *fiber.Ctx) (tenantID uuid.UUID, userID uuid.UUID, err error) {
	tenantIDStr, err := internalauth.GetTenantID(ctx)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}

	userIDStr, err := internalauth.GetUserID(ctx)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}

	tenantID, err = uuid.Parse(tenantIDStr)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}

	userID, err = uuid.Parse(userIDStr)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}

	return tenantID, userID, nil
}
