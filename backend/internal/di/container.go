package di

import (
	"database/sql"

	"flowforge-automation-backend/internal/auth"
	"flowforge-automation-backend/internal/config"
	"flowforge-automation-backend/internal/db"
	"flowforge-automation-backend/internal/websocket"
	authcontroller "flowforge-automation-backend/pkg/controller/auth"
	healthcontroller "flowforge-automation-backend/pkg/controller/health"
	runcontroller "flowforge-automation-backend/pkg/controller/run"
	usercontroller "flowforge-automation-backend/pkg/controller/user"
	workflowcontroller "flowforge-automation-backend/pkg/controller/workflow"
	authrepository "flowforge-automation-backend/pkg/repository/auth"
	healthrepository "flowforge-automation-backend/pkg/repository/health"
	runrepository "flowforge-automation-backend/pkg/repository/run"
	userrepository "flowforge-automation-backend/pkg/repository/user"
	versionrepository "flowforge-automation-backend/pkg/repository/version"
	workflowrepository "flowforge-automation-backend/pkg/repository/workflow"
	authservice "flowforge-automation-backend/pkg/service/auth"
	"flowforge-automation-backend/pkg/service/execution"
	healthservice "flowforge-automation-backend/pkg/service/health"
	runservice "flowforge-automation-backend/pkg/service/run"
	userservice "flowforge-automation-backend/pkg/service/user"
	workflowservice "flowforge-automation-backend/pkg/service/workflow"

	"gorm.io/gorm"
)

type Container struct {
	DB                 *sql.DB
	GormDB             *gorm.DB
	HealthController   *healthcontroller.Controller
	AuthController     *authcontroller.Controller
	WorkflowController *workflowcontroller.Controller
	RunController      *runcontroller.Controller
	UserController     *usercontroller.Controller
	AuthService        authservice.Service
	JWTManager         *auth.JWTManager
	WSHub              *websocket.Hub
}

func NewContainer(cfg config.Config) (*Container, error) {
	database, err := db.ConnectDatabase(cfg.DBDSN)
	if err != nil {
		return nil, err
	}

	gormDB, err := db.ConnectDatabaseGORM(cfg.DBDSN)
	if err != nil {
		_ = database.Close()
		return nil, err
	}

	// health
	healthRepo := healthrepository.NewHealthRepository(database)
	healthSvc := healthservice.NewHealthService(healthRepo)
	healthCtrl := healthcontroller.NewController(healthSvc)

	// auth
	authRepo := authrepository.NewAuthRepository(gormDB)
	authSvc := authservice.NewAuthService(authRepo, cfg.JWTSecret, cfg.TokenTTL)
	authCtrl := authcontroller.NewAuthController(authSvc)
	jwtMgr := auth.NewJWTManager(cfg.JWTSecret)

	// websocket hub
	wsHub := websocket.NewHub()

	// repositories
	workflowRepo := workflowrepository.NewWorkflowRepository(gormDB)
	versionRepo := versionrepository.NewVersionRepository(gormDB)
	runRepo := runrepository.NewRunRepository(gormDB)
	userRepo := userrepository.NewUserRepository(gormDB)

	// workflow service (now with version repo)
	workflowSvc := workflowservice.NewWorkflowService(workflowRepo, versionRepo)
	workflowCtrl := workflowcontroller.NewWorkflowController(workflowSvc)

	// execution engine
	engine := execution.NewEngine(wsHub, runRepo)

	// run service
	runSvc := runservice.NewRunService(runRepo, workflowRepo, engine)
	runCtrl := runcontroller.NewRunController(runSvc)

	// user service
	userSvc := userservice.NewUserService(userRepo)
	userCtrl := usercontroller.NewUserController(userSvc)

	return &Container{
		DB:                 database,
		GormDB:             gormDB,
		HealthController:   healthCtrl,
		AuthController:     authCtrl,
		WorkflowController: workflowCtrl,
		RunController:      runCtrl,
		UserController:     userCtrl,
		AuthService:        authSvc,
		JWTManager:         jwtMgr,
		WSHub:              wsHub,
	}, nil
}

func (c *Container) Close() error {
	if c.DB == nil {
		return nil
	}
	return c.DB.Close()
}
