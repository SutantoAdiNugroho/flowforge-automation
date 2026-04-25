package di

import (
	"database/sql"

	"flowforge-automation-backend/internal/auth"
	"flowforge-automation-backend/internal/config"
	"flowforge-automation-backend/internal/db"
	authcontroller "flowforge-automation-backend/pkg/controller/auth"
	healthcontroller "flowforge-automation-backend/pkg/controller/health"
	authrepository "flowforge-automation-backend/pkg/repository/auth"
	healthrepository "flowforge-automation-backend/pkg/repository/health"
	authservice "flowforge-automation-backend/pkg/service/auth"
	healthservice "flowforge-automation-backend/pkg/service/health"
	"gorm.io/gorm"
)

type Container struct {
	DB                *sql.DB
	GormDB            *gorm.DB
	HealthController  *healthcontroller.Controller
	AuthController    *authcontroller.Controller
	AuthService       authservice.Service
	JWTManager        *auth.JWTManager
}

func NewContainer(cfg config.Config) (*Container, error) {
	// Connect with sql.DB for health check
	database, err := db.ConnectDatabase(cfg.DBDSN)
	if err != nil {
		return nil, err
	}

	// Connect with GORM for ORM operations
	gormDB, err := db.ConnectDatabaseGORM(cfg.DBDSN)
	if err != nil {
		_ = database.Close()
		return nil, err
	}

	// Health check
	healthRepo := healthrepository.NewHealthRepository(database)
	healthSvc := healthservice.NewHealthService(healthRepo)
	healthCtrl := healthcontroller.NewController(healthSvc)

	// Auth layer
	authRepo := authrepository.NewAuthRepository(gormDB)
	authSvc := authservice.NewAuthService(authRepo, cfg.JWTSecret, cfg.TokenTTL)
	authCtrl := authcontroller.NewAuthController(authSvc)
	jwtMgr := auth.NewJWTManager(cfg.JWTSecret)

	return &Container{
		DB:               database,
		GormDB:           gormDB,
		HealthController: healthCtrl,
		AuthController:   authCtrl,
		AuthService:      authSvc,
		JWTManager:       jwtMgr,
	}, nil
}

func (c *Container) Close() error {
	if c.DB == nil {
		return nil
	}
	return c.DB.Close()
}
