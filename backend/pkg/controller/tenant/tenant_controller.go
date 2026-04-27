package tenant

import (
	"errors"
	"net/http"
	"strconv"

	"flowforge-automation-backend/pkg/model/dto"
	"flowforge-automation-backend/pkg/model/dto/response"
	tenantservice "flowforge-automation-backend/pkg/service/tenant"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Controller struct {
	tenantService tenantservice.Service
}

func NewTenantController(tenantService tenantservice.Service) *Controller {
	return &Controller{tenantService: tenantService}
}

func (c *Controller) Create(ctx *fiber.Ctx) error {
	req := &dto.CreateTenantRequest{}
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid request body"})
	}

	tenant, user, err := c.tenantService.Create(ctx.Context(), req)
	if err != nil {
		if errors.Is(err, tenantservice.ErrInvalidCreateRequest) {
			return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: err.Error()})
		}
		if errors.Is(err, tenantservice.ErrEmailAlreadyExists) {
			return ctx.Status(http.StatusConflict).JSON(dto.ErrorResponse{Error: "EMAIL_EXISTS", Message: err.Error()})
		}
		if errors.Is(err, tenantservice.ErrTenantSlugExists) {
			return ctx.Status(http.StatusConflict).JSON(dto.ErrorResponse{Error: "SLUG_EXISTS", Message: err.Error()})
		}
		if errors.Is(err, tenantservice.ErrInvalidTenantSlug) {
			return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_SLUG", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "TENANT_CREATE_FAILED", Message: "failed to create tenant"})
	}

	return ctx.Status(http.StatusCreated).JSON(fiber.Map{
		"tenant": tenant,
		"admin_user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (c *Controller) List(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size", "20"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	tenants, total, err := c.tenantService.List(ctx.Context(), pageSize, offset)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "TENANT_LIST_FAILED", Message: "failed to list tenants"})
	}

	return ctx.Status(http.StatusOK).JSON(response.NewPaginationResponse(page, pageSize, int(total), tenants))
}

func (c *Controller) GetByID(ctx *fiber.Ctx) error {
	tenantID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid tenant id"})
	}

	tenant, err := c.tenantService.GetByID(ctx.Context(), tenantID)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "TENANT_GET_FAILED", Message: "failed to get tenant"})
	}
	if tenant == nil {
		return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "TENANT_NOT_FOUND", Message: "tenant not found"})
	}

	return ctx.Status(http.StatusOK).JSON(tenant)
}

func (c *Controller) Update(ctx *fiber.Ctx) error {
	tenantID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid tenant id"})
	}

	req := &dto.UpdateTenantRequest{}
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid request body"})
	}

	tenant, err := c.tenantService.Update(ctx.Context(), tenantID, req)
	if err != nil {
		if errors.Is(err, tenantservice.ErrTenantNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "TENANT_NOT_FOUND", Message: err.Error()})
		}
		if errors.Is(err, tenantservice.ErrInvalidUpdateRequest) {
			return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "TENANT_UPDATE_FAILED", Message: "failed to update tenant"})
	}

	return ctx.Status(http.StatusOK).JSON(tenant)
}

func (c *Controller) Delete(ctx *fiber.Ctx) error {
	tenantID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid tenant id"})
	}

	err = c.tenantService.Delete(ctx.Context(), tenantID)
	if err != nil {
		if errors.Is(err, tenantservice.ErrTenantNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "TENANT_NOT_FOUND", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "TENANT_DELETE_FAILED", Message: "failed to delete tenant"})
	}

	return ctx.Status(http.StatusNoContent).Send(nil)
}
