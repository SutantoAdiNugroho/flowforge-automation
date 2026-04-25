package routes

import (
	healthcontroller "flowforge-automation-backend/pkg/controller/health"

	"github.com/gofiber/fiber/v2"
)

func Setup(healthController *healthcontroller.Controller) *fiber.App {
	app := fiber.New()

	api := app.Group("/api")
	api.Get("/health", healthController.GetHealth)

	return app
}
