package workflow

import (
	"errors"
	"net/http"
	"strconv"

	internalauth "flowforge-automation-backend/internal/auth"
	"flowforge-automation-backend/pkg/model/dto"
	"flowforge-automation-backend/pkg/model/dto/response"
	workflowservice "flowforge-automation-backend/pkg/service/workflow"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Controller struct {
	workflowService workflowservice.Service
}

func NewWorkflowController(workflowService workflowservice.Service) *Controller {
	return &Controller{workflowService: workflowService}
}

func (c *Controller) Create(ctx *fiber.Ctx) error {
	tenantID, userID, err := getAuthContextIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	req := &dto.CreateWorkflowRequest{}
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid request body"})
	}

	workflow, err := c.workflowService.Create(ctx.Context(), tenantID, userID, req)
	if err != nil {
		if errors.Is(err, workflowservice.ErrInvalidCreateWorkflowRequest) {
			return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: err.Error()})
		}
		if errors.Is(err, workflowservice.ErrInvalidWorkflowDefinition) {
			return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_DEFINITION", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "WORKFLOW_CREATE_FAILED", Message: "failed to create workflow"})
	}

	return ctx.Status(http.StatusCreated).JSON(workflow)
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

	workflows, total, err := c.workflowService.ListByTenant(ctx.Context(), tenantID, pageSize, offset)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "WORKFLOW_LIST_FAILED", Message: "failed to list workflows"})
	}

	return ctx.Status(http.StatusOK).JSON(response.NewPaginationResponse(page, pageSize, int(total), workflows))
}

func (c *Controller) GetByID(ctx *fiber.Ctx) error {
	tenantID, _, err := getAuthContextIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	workflowID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid workflow id"})
	}

	workflow, err := c.workflowService.GetByID(ctx.Context(), tenantID, workflowID)
	if err != nil {
		if errors.Is(err, workflowservice.ErrWorkflowNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "WORKFLOW_NOT_FOUND", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "WORKFLOW_GET_FAILED", Message: "failed to get workflow"})
	}

	return ctx.Status(http.StatusOK).JSON(workflow)
}

func (c *Controller) Update(ctx *fiber.Ctx) error {
	tenantID, userID, err := getAuthContextIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	workflowID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid workflow id"})
	}

	req := &dto.UpdateWorkflowRequest{}
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid request body"})
	}

	workflow, err := c.workflowService.Update(ctx.Context(), tenantID, userID, workflowID, req)
	if err != nil {
		if errors.Is(err, workflowservice.ErrInvalidUpdateWorkflowRequest) {
			return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: err.Error()})
		}
		if errors.Is(err, workflowservice.ErrInvalidWorkflowDefinition) {
			return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_DEFINITION", Message: err.Error()})
		}
		if errors.Is(err, workflowservice.ErrWorkflowNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "WORKFLOW_NOT_FOUND", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "WORKFLOW_UPDATE_FAILED", Message: "failed to update workflow"})
	}

	return ctx.Status(http.StatusOK).JSON(workflow)
}

func (c *Controller) CreateVersion(ctx *fiber.Ctx) error {
	tenantID, userID, err := getAuthContextIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	workflowID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid workflow id"})
	}

	req := &dto.CreateWorkflowVersionRequest{}
	if err := ctx.BodyParser(req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid request body"})
	}

	workflow, err := c.workflowService.CreateVersion(ctx.Context(), tenantID, userID, workflowID, req)
	if err != nil {
		if errors.Is(err, workflowservice.ErrInvalidUpdateWorkflowRequest) {
			return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: err.Error()})
		}
		if errors.Is(err, workflowservice.ErrInvalidWorkflowDefinition) {
			return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_DEFINITION", Message: err.Error()})
		}
		if errors.Is(err, workflowservice.ErrWorkflowNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "WORKFLOW_NOT_FOUND", Message: err.Error()})
		}
		if errors.Is(err, workflowservice.ErrVersionNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "VERSION_NOT_FOUND", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "VERSION_CREATE_FAILED", Message: "failed to create workflow version"})
	}

	return ctx.Status(http.StatusCreated).JSON(workflow)
}

func (c *Controller) ActivateVersion(ctx *fiber.Ctx) error {
	tenantID, userID, err := getAuthContextIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	workflowID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid workflow id"})
	}

	version, err := strconv.Atoi(ctx.Params("version"))
	if err != nil || version < 1 {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid version number"})
	}

	workflow, err := c.workflowService.ActivateVersion(ctx.Context(), tenantID, userID, workflowID, version)
	if err != nil {
		if errors.Is(err, workflowservice.ErrWorkflowNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "WORKFLOW_NOT_FOUND", Message: err.Error()})
		}
		if errors.Is(err, workflowservice.ErrVersionNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "VERSION_NOT_FOUND", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "ACTIVATE_VERSION_FAILED", Message: "failed to activate workflow version"})
	}

	return ctx.Status(http.StatusOK).JSON(workflow)
}

func (c *Controller) Delete(ctx *fiber.Ctx) error {
	tenantID, _, err := getAuthContextIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	workflowID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid workflow id"})
	}

	err = c.workflowService.Delete(ctx.Context(), tenantID, workflowID)
	if err != nil {
		if errors.Is(err, workflowservice.ErrWorkflowNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "WORKFLOW_NOT_FOUND", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "WORKFLOW_DELETE_FAILED", Message: "failed to delete workflow"})
	}

	return ctx.Status(http.StatusNoContent).Send(nil)
}

func (c *Controller) ListVersions(ctx *fiber.Ctx) error {
	tenantID, _, err := getAuthContextIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	workflowID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid workflow id"})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size", "20"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	versions, total, err := c.workflowService.ListVersions(ctx.Context(), tenantID, workflowID, pageSize, offset)
	if err != nil {
		if errors.Is(err, workflowservice.ErrWorkflowNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "WORKFLOW_NOT_FOUND", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "VERSION_LIST_FAILED", Message: "failed to list versions"})
	}

	return ctx.Status(http.StatusOK).JSON(response.NewPaginationResponse(page, pageSize, int(total), versions))
}

func (c *Controller) Rollback(ctx *fiber.Ctx) error {
	tenantID, userID, err := getAuthContextIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	workflowID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid workflow id"})
	}

	version, err := strconv.Atoi(ctx.Params("version"))
	if err != nil || version < 1 {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid version number"})
	}

	workflow, err := c.workflowService.Rollback(ctx.Context(), tenantID, userID, workflowID, version)
	if err != nil {
		if errors.Is(err, workflowservice.ErrWorkflowNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "WORKFLOW_NOT_FOUND", Message: err.Error()})
		}
		if errors.Is(err, workflowservice.ErrVersionNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "VERSION_NOT_FOUND", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "ROLLBACK_FAILED", Message: "failed to rollback workflow"})
	}

	return ctx.Status(http.StatusOK).JSON(workflow)
}

func getAuthContextIDs(ctx *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	tenantIDStr, err := internalauth.GetTenantID(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	userIDStr, err := internalauth.GetUserID(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return tenantID, userID, nil
}
