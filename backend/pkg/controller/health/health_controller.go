package health

import (
	"context"
	"time"

	healthservice "flowforge-automation-backend/pkg/service/health"

	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	healthService healthservice.Service
}

func NewController(healthService healthservice.Service) *Controller {
	return &Controller{healthService: healthService}
}

func (c *Controller) GetHealth(ctx *fiber.Ctx) error {
	reqCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := c.healthService.Check(reqCtx)
	if err != nil {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(res)
	}

	return ctx.Status(fiber.StatusOK).JSON(res)
}
