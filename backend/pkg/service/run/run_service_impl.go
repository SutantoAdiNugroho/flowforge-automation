package run

import (
	"context"
	"errors"

	"flowforge-automation-backend/pkg/model/domain"
	"flowforge-automation-backend/pkg/model/dto"
	runrepo "flowforge-automation-backend/pkg/repository/run"
	workflowrepo "flowforge-automation-backend/pkg/repository/workflow"
	"flowforge-automation-backend/pkg/service/execution"
	"github.com/google/uuid"
)

var (
	ErrWorkflowNotFound = errors.New("workflow not found")
	ErrRunNotFound      = errors.New("run not found")
	ErrRunNotCancellable = errors.New("run is not in a cancellable state")
)

type service struct {
	runRepo      runrepo.Repository
	workflowRepo workflowrepo.Repository
	engine       *execution.Engine
}

func NewRunService(runRepo runrepo.Repository, workflowRepo workflowrepo.Repository, engine *execution.Engine) Service {
	return &service{
		runRepo:      runRepo,
		workflowRepo: workflowRepo,
		engine:       engine,
	}
}

func (s *service) TriggerRun(ctx context.Context, tenantID, userID, workflowID uuid.UUID, req *dto.TriggerRunRequest) (*domain.WorkflowRun, error) {
	wf, err := s.workflowRepo.GetByIDAndTenant(ctx, workflowID, tenantID)
	if err != nil {
		return nil, err
	}
	if wf == nil {
		return nil, ErrWorkflowNotFound
	}

	run := &domain.WorkflowRun{
		BaseModel:   domain.BaseModel{ID: uuid.New()},
		WorkflowID:  workflowID,
		TenantID:    tenantID,
		Status:      "pending",
		TriggeredBy: "manual",
	}

	if req != nil && req.Inputs != nil {
		run.Metadata = domain.JSONB(req.Inputs)
	}

	if err := s.runRepo.Create(ctx, run); err != nil {
		return nil, err
	}

	// execute in background
	go s.engine.Execute(context.Background(), wf, run)

	return run, nil
}

func (s *service) GetRun(ctx context.Context, tenantID, runID uuid.UUID) (*dto.RunDetailResponse, error) {
	run, err := s.runRepo.GetByIDAndTenant(ctx, runID, tenantID)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, ErrRunNotFound
	}

	steps, err := s.runRepo.GetStepsByRunID(ctx, runID)
	if err != nil {
		return nil, err
	}

	return &dto.RunDetailResponse{
		Run:   *run,
		Steps: steps,
	}, nil
}

func (s *service) ListRunsByWorkflow(ctx context.Context, tenantID, workflowID uuid.UUID, limit, offset int) ([]domain.WorkflowRun, int64, error) {
	return s.runRepo.ListByWorkflow(ctx, workflowID, tenantID, limit, offset)
}

func (s *service) ListRunsByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.WorkflowRun, int64, error) {
	return s.runRepo.ListByTenant(ctx, tenantID, limit, offset)
}

func (s *service) CancelRun(ctx context.Context, tenantID, runID uuid.UUID) error {
	run, err := s.runRepo.GetByIDAndTenant(ctx, runID, tenantID)
	if err != nil {
		return err
	}
	if run == nil {
		return ErrRunNotFound
	}
	if run.Status != "pending" && run.Status != "running" {
		return ErrRunNotCancellable
	}
	return s.runRepo.UpdateStatus(ctx, runID, "cancelled", "cancelled by user")
}

func (s *service) GetStats(ctx context.Context, tenantID uuid.UUID, hours int) (*runrepo.RunStats, error) {
	if hours <= 0 {
		hours = 24
	}
	return s.runRepo.GetRunStats(ctx, tenantID, hours)
}


