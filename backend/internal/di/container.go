package di

import (
	"database/sql"

	"flowforge-automation-backend/internal/config"
	"flowforge-automation-backend/internal/db"
	healthcontroller "flowforge-automation-backend/pkg/controller/health"
	healthrepository "flowforge-automation-backend/pkg/repository/health"
	healthservice "flowforge-automation-backend/pkg/service/health"
)

type Container struct {
	DB               *sql.DB
	HealthController *healthcontroller.Controller
}

func NewContainer(cfg config.Config) (*Container, error) {
	database, err := db.ConnectDatabase(cfg.DBDSN)
	if err != nil {
		return nil, err
	}

	healthRepo := healthrepository.NewHealthRepository(database)
	healthService := healthservice.NewHealthService(healthRepo)
	healthController := healthcontroller.NewController(healthService)

	return &Container{
		DB:               database,
		HealthController: healthController,
	}, nil
}

func (c *Container) Close() error {
	if c.DB == nil {
		return nil
	}
	return c.DB.Close()
}
