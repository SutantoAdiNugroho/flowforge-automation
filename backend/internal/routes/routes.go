package routes

import (
	"flowforge-automation-backend/internal/auth"
	"flowforge-automation-backend/internal/middleware"
	"flowforge-automation-backend/internal/websocket"
	authcontroller "flowforge-automation-backend/pkg/controller/auth"
	healthcontroller "flowforge-automation-backend/pkg/controller/health"
	runcontroller "flowforge-automation-backend/pkg/controller/run"
	workflowcontroller "flowforge-automation-backend/pkg/controller/workflow"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type Controllers struct {
	Health   *healthcontroller.Controller
	Auth     *authcontroller.Controller
	Workflow *workflowcontroller.Controller
	Run      *runcontroller.Controller
}

func Setup(ctrl Controllers, jwtManager *auth.JWTManager, wsHub *websocket.Hub) *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// rate limiter: 50 req/s burst 100
	rl := middleware.NewRateLimiter(50, 100)
	app.Use(rl.Middleware())

	api := app.Group("/api")

	// public
	api.Get("/health", ctrl.Health.GetHealth)
	api.Post("/auth/register", ctrl.Auth.Register)
	api.Post("/auth/login", ctrl.Auth.Login)

	// protected
	protected := api.Group("/")
	protected.Use(auth.Middleware(jwtManager))

	// sse stream for real-time run monitoring
	protected.Get("/events", websocket.SSEHandler(wsHub))

	// runs (all authenticated users)
	protected.Get("/runs", ctrl.Run.ListRunsByTenant)
	protected.Get("/runs/stats", ctrl.Run.GetStats)
	protected.Get("/runs/:runId", ctrl.Run.GetRun)

	// workflows - read (viewer+)
	viewerPlus := protected.Group("/")
	viewerPlus.Use(auth.RoleBasedMiddleware("admin", "editor", "viewer"))
	viewerPlus.Get("/workflows", ctrl.Workflow.List)
	viewerPlus.Get("/workflows/:id", ctrl.Workflow.GetByID)
	viewerPlus.Get("/workflows/:id/versions", ctrl.Workflow.ListVersions)
	viewerPlus.Get("/workflows/:id/runs", ctrl.Run.ListRunsByWorkflow)

	// workflows - write (editor+)
	editorPlus := protected.Group("/")
	editorPlus.Use(auth.RoleBasedMiddleware("admin", "editor"))
	editorPlus.Post("/workflows", ctrl.Workflow.Create)
	editorPlus.Put("/workflows/:id", ctrl.Workflow.Update)
	editorPlus.Post("/workflows/:id/runs", ctrl.Run.TriggerRun)
	editorPlus.Post("/runs/:runId/cancel", ctrl.Run.CancelRun)

	// workflows - admin only
	adminOnly := protected.Group("/")
	adminOnly.Use(auth.RoleBasedMiddleware("admin"))
	adminOnly.Delete("/workflows/:id", ctrl.Workflow.Delete)
	adminOnly.Put("/workflows/:id/rollback/:version", ctrl.Workflow.Rollback)

	return app
}
