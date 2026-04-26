package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"flowforge-automation-backend/pkg/model/domain"
	"flowforge-automation-backend/pkg/model/dto"
	"flowforge-automation-backend/pkg/service/execution"
	versionrepo "flowforge-automation-backend/pkg/repository/version"
	workflowrepository "flowforge-automation-backend/pkg/repository/workflow"
	"github.com/google/uuid"
)

type service struct {
	repo         workflowrepository.Repository
	versionRepo  versionrepo.Repository
	dagValidator *execution.DAGValidator
}

func NewWorkflowService(repo workflowrepository.Repository, versionRepo versionrepo.Repository) Service {
	return &service{
		repo:         repo,
		versionRepo:  versionRepo,
		dagValidator: execution.NewDAGValidator(),
	}
}

func (s *service) validateDefinition(definition domain.JSONB) error {
	validationResult, err := s.dagValidator.Validate(definition)
	if err != nil {
		return err
	}
	if !validationResult.Valid {
		return ErrInvalidWorkflowDefinition
	}
	return nil
}

func (s *service) isDAGDefinition(def domain.JSONB) bool {
	defBytes, err := json.Marshal(def)
	if err != nil {
		return false
	}
	var dag domain.DAGDefinition
	if err := json.Unmarshal(defBytes, &dag); err != nil {
		return false
	}
	return len(dag.Steps) > 0
}

func (s *service) Create(ctx context.Context, tenantID, userID uuid.UUID, req *dto.CreateWorkflowRequest) (*domain.Workflow, error) {
	if req == nil || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.TriggerType) == "" || len(req.Definition) == 0 {
		return nil, ErrInvalidCreateWorkflowRequest
	}

	triggerType := strings.ToLower(strings.TrimSpace(req.TriggerType))
	if triggerType != "manual" && triggerType != "cron" && triggerType != "webhook" {
		return nil, ErrInvalidCreateWorkflowRequest
	}

	if triggerType == "cron" && strings.TrimSpace(req.CronExpression) == "" {
		return nil, ErrInvalidCreateWorkflowRequest
	}

	if s.isDAGDefinition(req.Definition) {
		if err := s.validateDefinition(req.Definition); err != nil {
			return nil, err
		}
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	workflow := &domain.Workflow{
		BaseModel:      domain.BaseModel{ID: uuid.New()},
		TenantID:       tenantID,
		Name:           strings.TrimSpace(req.Name),
		Description:    strings.TrimSpace(req.Description),
		Definition:     req.Definition,
		Version:        1,
		IsActive:       isActive,
		TriggerType:    triggerType,
		CronExpression: strings.TrimSpace(req.CronExpression),
		CreatedByID:    userID,
	}

	if err := s.repo.Create(ctx, workflow); err != nil {
		return nil, err
	}

	// save initial version
	s.saveVersion(ctx, workflow.ID, 1, req.Definition, userID)

	return workflow, nil
}

func (s *service) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Workflow, int64, error) {
	return s.repo.ListByTenant(ctx, tenantID, limit, offset)
}

func (s *service) GetByID(ctx context.Context, tenantID, workflowID uuid.UUID) (*domain.Workflow, error) {
	workflow, err := s.repo.GetByIDAndTenant(ctx, workflowID, tenantID)
	if err != nil {
		return nil, err
	}
	if workflow == nil {
		return nil, ErrWorkflowNotFound
	}
	return workflow, nil
}

func (s *service) Update(ctx context.Context, tenantID, userID, workflowID uuid.UUID, req *dto.UpdateWorkflowRequest) (*domain.Workflow, error) {
	if req == nil || strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.TriggerType) == "" || len(req.Definition) == 0 {
		return nil, ErrInvalidUpdateWorkflowRequest
	}

	triggerType := strings.ToLower(strings.TrimSpace(req.TriggerType))
	if triggerType != "manual" && triggerType != "cron" && triggerType != "webhook" {
		return nil, ErrInvalidUpdateWorkflowRequest
	}

	if triggerType == "cron" && strings.TrimSpace(req.CronExpression) == "" {
		return nil, ErrInvalidUpdateWorkflowRequest
	}

	if s.isDAGDefinition(req.Definition) {
		if err := s.validateDefinition(req.Definition); err != nil {
			return nil, err
		}
	}

	existing, err := s.repo.GetByIDAndTenant(ctx, workflowID, tenantID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrWorkflowNotFound
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	newVersion := existing.Version + 1

	workflow := &domain.Workflow{
		BaseModel:      existing.BaseModel,
		TenantID:       tenantID,
		Name:           strings.TrimSpace(req.Name),
		Description:    strings.TrimSpace(req.Description),
		Definition:     req.Definition,
		Version:        newVersion,
		IsActive:       isActive,
		TriggerType:    triggerType,
		CronExpression: strings.TrimSpace(req.CronExpression),
		CreatedByID:    existing.CreatedByID,
	}
	workflow.ID = workflowID
	workflow.UpdatedBy = &userID

	if err := s.repo.Update(ctx, workflow, tenantID); err != nil {
		return nil, err
	}

	// persist version snapshot
	s.saveVersion(ctx, workflowID, newVersion, req.Definition, userID)

	return workflow, nil
}

func (s *service) Delete(ctx context.Context, tenantID, workflowID uuid.UUID) error {
	existing, err := s.repo.GetByIDAndTenant(ctx, workflowID, tenantID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrWorkflowNotFound
	}

	return s.repo.Delete(ctx, workflowID, tenantID)
}

func (s *service) ListVersions(ctx context.Context, tenantID, workflowID uuid.UUID, limit, offset int) ([]domain.WorkflowVersion, int64, error) {
	// verify workflow belongs to tenant
	wf, err := s.repo.GetByIDAndTenant(ctx, workflowID, tenantID)
	if err != nil {
		return nil, 0, err
	}
	if wf == nil {
		return nil, 0, ErrWorkflowNotFound
	}

	return s.versionRepo.ListByWorkflow(ctx, workflowID, limit, offset)
}

func (s *service) Rollback(ctx context.Context, tenantID, userID, workflowID uuid.UUID, version int) (*domain.Workflow, error) {
	existing, err := s.repo.GetByIDAndTenant(ctx, workflowID, tenantID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrWorkflowNotFound
	}

	ver, err := s.versionRepo.GetByWorkflowAndVersion(ctx, workflowID, version)
	if err != nil {
		return nil, err
	}
	if ver == nil {
		return nil, ErrVersionNotFound
	}

	newVersion := existing.Version + 1

	existing.Definition = ver.Definition
	existing.Version = newVersion
	existing.UpdatedBy = &userID

	if err := s.repo.Update(ctx, existing, tenantID); err != nil {
		return nil, err
	}

	s.saveVersion(ctx, workflowID, newVersion, ver.Definition, userID)

	return existing, nil
}

func (s *service) saveVersion(ctx context.Context, workflowID uuid.UUID, version int, definition domain.JSONB, userID uuid.UUID) {
	v := &domain.WorkflowVersion{
		ID:          uuid.New(),
		WorkflowID:  workflowID,
		Version:     version,
		Definition:  definition,
		CreatedByID: userID,
	}
	_ = s.versionRepo.Create(ctx, v)
}

var ErrVersionNotFound = errors.New("version not found")
