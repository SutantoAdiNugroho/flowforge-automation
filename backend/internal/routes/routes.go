package routes

import (
	"flowforge-automation-backend/internal/auth"
	authcontroller "flowforge-automation-backend/pkg/controller/auth"
	healthcontroller "flowforge-automation-backend/pkg/controller/health"

	"github.com/gofiber/fiber/v2"
)

func Setup(healthController *healthcontroller.Controller, authController *authcontroller.Controller, jwtManager *auth.JWTManager) *fiber.App {
	app := fiber.New()

	api := app.Group("/api")

	// Public endpoints
	api.Get("/health", healthController.GetHealth)
	api.Post("/auth/register", authController.Register)
	api.Post("/auth/login", authController.Login)

	// Protected endpoints
	protected := api.Group("/")
	protected.Use(auth.Middleware(jwtManager))
	// Add protected routes here as needed

	return app
}
