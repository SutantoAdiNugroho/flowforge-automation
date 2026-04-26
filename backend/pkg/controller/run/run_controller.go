package run

import (
	"errors"
	"net/http"
	"strconv"

	internalauth "flowforge-automation-backend/internal/auth"
	"flowforge-automation-backend/pkg/model/dto"
	"flowforge-automation-backend/pkg/model/dto/response"
	runservice "flowforge-automation-backend/pkg/service/run"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Controller struct {
	runService runservice.Service
}

func NewRunController(runService runservice.Service) *Controller {
	return &Controller{runService: runService}
}

func (c *Controller) TriggerRun(ctx *fiber.Ctx) error {
	tenantID, userID, err := getIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	workflowID, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid workflow id"})
	}

	req := &dto.TriggerRunRequest{}
	if ctx.Body() != nil && len(ctx.Body()) > 0 {
		_ = ctx.BodyParser(req)
	}

	run, err := c.runService.TriggerRun(ctx.Context(), tenantID, userID, workflowID, req)
	if err != nil {
		if errors.Is(err, runservice.ErrWorkflowNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "WORKFLOW_NOT_FOUND", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "TRIGGER_FAILED", Message: "failed to trigger run"})
	}

	return ctx.Status(http.StatusCreated).JSON(run)
}

func (c *Controller) GetRun(ctx *fiber.Ctx) error {
	tenantID, _, err := getIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	runID, err := uuid.Parse(ctx.Params("runId"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid run id"})
	}

	detail, err := c.runService.GetRun(ctx.Context(), tenantID, runID)
	if err != nil {
		if errors.Is(err, runservice.ErrRunNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "RUN_NOT_FOUND", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "RUN_GET_FAILED", Message: "failed to get run"})
	}

	return ctx.Status(http.StatusOK).JSON(detail)
}

func (c *Controller) ListRunsByWorkflow(ctx *fiber.Ctx) error {
	tenantID, _, err := getIDs(ctx)
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
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	runs, total, err := c.runService.ListRunsByWorkflow(ctx.Context(), tenantID, workflowID, pageSize, offset)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "RUN_LIST_FAILED", Message: "failed to list runs"})
	}

	return ctx.Status(http.StatusOK).JSON(response.NewPaginationResponse(page, pageSize, int(total), runs))
}

func (c *Controller) ListRunsByTenant(ctx *fiber.Ctx) error {
	tenantID, _, err := getIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	runs, total, err := c.runService.ListRunsByTenant(ctx.Context(), tenantID, pageSize, offset)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "RUN_LIST_FAILED", Message: "failed to list runs"})
	}

	return ctx.Status(http.StatusOK).JSON(response.NewPaginationResponse(page, pageSize, int(total), runs))
}

func (c *Controller) CancelRun(ctx *fiber.Ctx) error {
	tenantID, _, err := getIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	runID, err := uuid.Parse(ctx.Params("runId"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(dto.ErrorResponse{Error: "INVALID_REQUEST", Message: "invalid run id"})
	}

	err = c.runService.CancelRun(ctx.Context(), tenantID, runID)
	if err != nil {
		if errors.Is(err, runservice.ErrRunNotFound) {
			return ctx.Status(http.StatusNotFound).JSON(dto.ErrorResponse{Error: "RUN_NOT_FOUND", Message: err.Error()})
		}
		if errors.Is(err, runservice.ErrRunNotCancellable) {
			return ctx.Status(http.StatusConflict).JSON(dto.ErrorResponse{Error: "RUN_NOT_CANCELLABLE", Message: err.Error()})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "CANCEL_FAILED", Message: "failed to cancel run"})
	}

	return ctx.Status(http.StatusOK).JSON(map[string]string{"status": "cancelled"})
}

func (c *Controller) GetStats(ctx *fiber.Ctx) error {
	tenantID, _, err := getIDs(ctx)
	if err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(dto.ErrorResponse{Error: "UNAUTHORIZED", Message: "unauthorized"})
	}

	hours, _ := strconv.Atoi(ctx.Query("hours", "24"))

	stats, err := c.runService.GetStats(ctx.Context(), tenantID, hours)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{Error: "STATS_FAILED", Message: "failed to get stats"})
	}

	return ctx.Status(http.StatusOK).JSON(stats)
}

func getIDs(ctx *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
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
