package main

import (
	"log"

	"flowforge-automation-backend/internal/config"
	"flowforge-automation-backend/internal/db"
	"flowforge-automation-backend/internal/di"
	"flowforge-automation-backend/internal/routes"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using env variables")
	}

	cfg := config.Load()
	container, err := di.NewContainer(cfg)
	if err != nil {
		log.Fatalf("failed to initialize app container: %v", err)
	}
	defer container.Close()

	if cfg.MigrateOnStart {
		if err := db.RunMigrations(cfg.DBDSN, "file://migrations"); err != nil {
			log.Fatalf("failed to run migrations: %v", err)
		}
	}

	// Start scheduler
	container.Scheduler.Start()
	defer container.Scheduler.Stop()

	ctrl := routes.Controllers{
		Health:   container.HealthController,
		Auth:     container.AuthController,
		Workflow: container.WorkflowController,
		Run:      container.RunController,
		User:     container.UserController,
	}

	app := routes.Setup(ctrl, container.JWTManager, container.WSHub)
	log.Printf("flowforge backend running on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("failed to run backend: %v", err)
	}
}
