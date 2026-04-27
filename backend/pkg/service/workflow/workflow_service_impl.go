package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"flowforge-automation-backend/pkg/model/domain"
	"flowforge-automation-backend/pkg/model/dto"
	versionrepo "flowforge-automation-backend/pkg/repository/version"
	workflowrepository "flowforge-automation-backend/pkg/repository/workflow"
	"flowforge-automation-backend/pkg/service/execution"

	"github.com/google/uuid"
)

type service struct {
	repo         workflowrepository.Repository
	versionRepo  versionrepo.Repository
	dagValidator *execution.DAGValidator
	scheduler    interface{ SyncWorkflows() error }
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

	s.saveVersion(ctx, workflow, userID)

	if s.scheduler != nil {
		_ = s.scheduler.SyncWorkflows()
	}

	return workflow, nil
}

func (s *service) nextVersionNumber(ctx context.Context, workflowID uuid.UUID) (int, error) {
	latest, err := s.versionRepo.GetLatestByWorkflow(ctx, workflowID)
	if err != nil {
		return 0, err
	}
	if latest == nil {
		return 1, nil
	}
	return latest.Version + 1, nil
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

	contentChanged := strings.TrimSpace(existing.Name) != strings.TrimSpace(req.Name) ||
		strings.TrimSpace(existing.Description) != strings.TrimSpace(req.Description) ||
		strings.ToLower(strings.TrimSpace(existing.TriggerType)) != triggerType ||
		strings.TrimSpace(existing.CronExpression) != strings.TrimSpace(req.CronExpression) ||
		!reflect.DeepEqual(existing.Definition, req.Definition)

	newVersion := existing.Version
	if contentChanged {
		nextVersion, err := s.nextVersionNumber(ctx, workflowID)
		if err != nil {
			return nil, err
		}
		newVersion = nextVersion
	}

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

	if contentChanged {
		s.saveVersion(ctx, workflow, userID)
	}

	if s.scheduler != nil {
		_ = s.scheduler.SyncWorkflows()
	}

	return workflow, nil
}

func (s *service) CreateVersion(ctx context.Context, tenantID, userID, workflowID uuid.UUID, req *dto.CreateWorkflowVersionRequest) (*domain.Workflow, error) {
	if req == nil {
		return nil, ErrInvalidUpdateWorkflowRequest
	}

	existing, err := s.repo.GetByIDAndTenant(ctx, workflowID, tenantID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrWorkflowNotFound
	}

	baseName := existing.Name
	baseDescription := existing.Description
	baseTriggerType := existing.TriggerType
	baseCronExpression := existing.CronExpression
	baseDefinition := existing.Definition

	if req.ImportFromVersion != nil {
		ver, err := s.versionRepo.GetByWorkflowAndVersion(ctx, workflowID, *req.ImportFromVersion)
		if err != nil {
			return nil, err
		}
		if ver == nil {
			return nil, ErrVersionNotFound
		}
		baseName = ver.Name
		baseDescription = ver.Description
		baseTriggerType = ver.TriggerType
		baseCronExpression = ver.CronExpression
		baseDefinition = ver.Definition
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = strings.TrimSpace(baseName)
	}

	description := strings.TrimSpace(req.Description)
	if description == "" {
		description = strings.TrimSpace(baseDescription)
	}

	triggerType := strings.ToLower(strings.TrimSpace(req.TriggerType))
	if triggerType == "" {
		triggerType = strings.ToLower(strings.TrimSpace(baseTriggerType))
	}

	cronExpression := strings.TrimSpace(req.CronExpression)
	if cronExpression == "" {
		cronExpression = strings.TrimSpace(baseCronExpression)
	}

	definition := req.Definition
	if len(definition) == 0 {
		definition = baseDefinition
	}

	if name == "" || triggerType == "" || len(definition) == 0 {
		return nil, ErrInvalidUpdateWorkflowRequest
	}

	if triggerType != "manual" && triggerType != "cron" && triggerType != "webhook" {
		return nil, ErrInvalidUpdateWorkflowRequest
	}

	if triggerType == "cron" && cronExpression == "" {
		return nil, ErrInvalidUpdateWorkflowRequest
	}

	if s.isDAGDefinition(definition) {
		if err := s.validateDefinition(definition); err != nil {
			return nil, err
		}
	}

	newVersion, err := s.nextVersionNumber(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	existing.Name = name
	existing.Description = description
	existing.TriggerType = triggerType
	existing.CronExpression = cronExpression
	existing.Definition = definition
	existing.Version = newVersion
	existing.UpdatedBy = &userID

	if err := s.repo.Update(ctx, existing, tenantID); err != nil {
		return nil, err
	}

	s.saveVersion(ctx, existing, userID)

	if s.scheduler != nil {
		_ = s.scheduler.SyncWorkflows()
	}

	return existing, nil
}

func (s *service) Delete(ctx context.Context, tenantID, workflowID uuid.UUID) error {
	existing, err := s.repo.GetByIDAndTenant(ctx, workflowID, tenantID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrWorkflowNotFound
	}

	if err := s.repo.Delete(ctx, workflowID, tenantID); err != nil {
		return err
	}

	if s.scheduler != nil {
		_ = s.scheduler.SyncWorkflows()
	}

	return nil
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

func (s *service) ActivateVersion(ctx context.Context, tenantID, userID, workflowID uuid.UUID, version int) (*domain.Workflow, error) {
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

	existing.Name = ver.Name
	existing.Description = ver.Description
	existing.TriggerType = ver.TriggerType
	existing.CronExpression = ver.CronExpression
	existing.Definition = ver.Definition
	existing.Version = ver.Version
	existing.UpdatedBy = &userID

	if err := s.repo.Update(ctx, existing, tenantID); err != nil {
		return nil, err
	}

	if s.scheduler != nil {
		_ = s.scheduler.SyncWorkflows()
	}

	return existing, nil
}

func (s *service) Rollback(ctx context.Context, tenantID, userID, workflowID uuid.UUID, version int) (*domain.Workflow, error) {
	return s.ActivateVersion(ctx, tenantID, userID, workflowID, version)
}

func (s *service) saveVersion(ctx context.Context, workflow *domain.Workflow, userID uuid.UUID) {
	v := &domain.WorkflowVersion{
		ID:             uuid.New(),
		WorkflowID:     workflow.ID,
		Version:        workflow.Version,
		Name:           workflow.Name,
		Description:    workflow.Description,
		TriggerType:    workflow.TriggerType,
		CronExpression: workflow.CronExpression,
		Definition:     workflow.Definition,
		CreatedByID:    userID,
	}
	_ = s.versionRepo.Create(ctx, v)
}

func (s *service) SetScheduler(scheduler interface{ SyncWorkflows() error }) {
	s.scheduler = scheduler
}

var ErrVersionNotFound = errors.New("version not found")
