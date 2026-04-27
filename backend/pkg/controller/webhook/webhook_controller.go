package webhook

import (
	"flowforge-automation-backend/pkg/model/dto"
	"flowforge-automation-backend/pkg/repository/workflow"
	"flowforge-automation-backend/pkg/service/run"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Controller struct {
	workflowRepo workflow.Repository
	runSvc       run.Service
}

func NewController(workflowRepo workflow.Repository, runSvc run.Service) *Controller {
	return &Controller{
		workflowRepo: workflowRepo,
		runSvc:       runSvc,
	}
}

func (c *Controller) HandleWebhook(ctx *fiber.Ctx) error {
	workflowIDStr := ctx.Params("workflowId")
	workflowID, err := uuid.Parse(workflowIDStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid workflow id"})
	}

	wf, err := c.workflowRepo.GetByIDPublic(ctx.Context(), workflowID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch workflow"})
	}
	if wf == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "workflow not found"})
	}

	if wf.TriggerType != "webhook" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "workflow is not configured for webhook trigger"})
	}

	if !wf.IsActive {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "workflow is inactive"})
	}

	// Inputs from request body
	var inputs map[string]interface{}
	if err := ctx.BodyParser(&inputs); err != nil {
		// If body is empty or not JSON, just continue with empty inputs
		inputs = make(map[string]interface{})
	}

	// Trigger run using the workflow's tenant and creator
	run, err := c.runSvc.TriggerRun(ctx.Context(), wf.TenantID, wf.CreatedByID, wf.ID, &dto.TriggerRunRequest{
		TriggeredBy: "webhook",
		Inputs:      inputs,
	})

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "workflow triggered",
		"run_id":  run.ID,
	})
}
