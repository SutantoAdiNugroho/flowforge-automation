package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"flowforge-automation-backend/pkg/model/domain"

	"github.com/google/uuid"
)

type StepUpdateEvent struct {
	RunID     uuid.UUID              `json:"run_id"`
	StepID    string                 `json:"step_id"`
	StepName  string                 `json:"step_name"`
	Status    string                 `json:"status"`
	Error     string                 `json:"error,omitempty"`
	Output    map[string]interface{} `json:"output,omitempty"`
	Retry     int                    `json:"retry_count"`
	Timestamp time.Time              `json:"timestamp"`
}

type RunUpdateEvent struct {
	RunID     uuid.UUID `json:"run_id"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type EventBroadcaster interface {
	BroadcastStepUpdate(event StepUpdateEvent)
	BroadcastRunUpdate(event RunUpdateEvent)
}

type RunPersister interface {
	UpdateStatus(ctx context.Context, runID uuid.UUID, status string, errorMsg string) error
	CreateStepExecution(ctx context.Context, step *domain.StepExecution) error
	UpdateStepExecution(ctx context.Context, step *domain.StepExecution) error
}

type Engine struct {
	broadcaster EventBroadcaster
	persister   RunPersister
	retry       *retryExecutor
}

func NewEngine(broadcaster EventBroadcaster, persister RunPersister) *Engine {
	return &Engine{
		broadcaster: broadcaster,
		persister:   persister,
		retry:       newRetryExecutor(),
	}
}

func (e *Engine) Execute(ctx context.Context, workflow *domain.Workflow, run *domain.WorkflowRun) {
	dag, err := e.parseDefinition(workflow.Definition)
	if err != nil {
		e.failRun(ctx, run.ID, fmt.Sprintf("failed to parse definition: %v", err))
		return
	}

	levels, err := domain.TopologicalLevels(dag.Steps)
	if err != nil || len(levels) == 0 {
		e.failRun(ctx, run.ID, "failed to compute execution order")
		return
	}

	// global timeout
	timeout := 10 * time.Minute
	if dag.Timeout > 0 {
		timeout = time.Duration(dag.Timeout) * time.Millisecond
	}
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	now := time.Now()
	run.StartedAt = &now
	run.Status = "running"
	_ = e.persister.UpdateStatus(execCtx, run.ID, "running", "")
	e.broadcastRunUpdate(run.ID, "running", "")

	// shared state: step outputs keyed by step id
	outputs := &sync.Map{}
	skipped := &sync.Map{}

	for _, level := range levels {
		if execCtx.Err() != nil {
			e.failRun(ctx, run.ID, "workflow timeout exceeded")
			return
		}

		var wg sync.WaitGroup
		errCh := make(chan error, len(level))

		for _, step := range level {
			wg.Add(1)
			go func(s domain.DAGStep) {
				defer wg.Done()

				// check if any dependency was skipped or condition returned false
				if e.shouldSkip(s, outputs, skipped) {
					skipped.Store(s.ID, true)
					e.recordStep(execCtx, run.ID, s, "skipped", nil, "", 0)
					return
				}

				err := e.executeStep(execCtx, run.ID, s, outputs, skipped)
				if err != nil {
					errCh <- fmt.Errorf("step %s failed: %w", s.ID, err)
				}
			}(step)
		}

		wg.Wait()
		close(errCh)

		for stepErr := range errCh {
			e.failRun(ctx, run.ID, stepErr.Error())
			return
		}
	}

	completed := time.Now()
	run.CompletedAt = &completed
	run.Status = "success"
	_ = e.persister.UpdateStatus(ctx, run.ID, "success", "")
	e.broadcastRunUpdate(run.ID, "success", "")
}

func (e *Engine) executeStep(
	ctx context.Context,
	runID uuid.UUID,
	step domain.DAGStep,
	outputs *sync.Map,
	skipped *sync.Map,
) error {
	stepTimeout := 60 * time.Second
	if step.Timeout > 0 {
		stepTimeout = time.Duration(step.Timeout) * time.Millisecond
	}
	stepCtx, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()

	stepExec := e.createStepRecord(runID, step)
	if err := e.persister.CreateStepExecution(ctx, stepExec); err != nil {
		fmt.Printf("failed to create step execution: %v\n", err)
		return fmt.Errorf("failed to create step: %w", err)
	}
	e.broadcastStepUpdate(runID, step, "running", "", nil, 0)

	inputData := e.collectInputs(step, outputs)
	runner := getRunner(step.Type)

	result, retries, err := e.retry.executeWithRetry(stepCtx, step.RetryPolicy, func(c context.Context) (map[string]interface{}, error) {
		return runner.Run(c, step, inputData)
	})

	now := time.Now()
	stepExec.CompletedAt = &now
	stepExec.RetryCount = retries
	dur := int(now.Sub(*stepExec.StartedAt).Milliseconds())
	stepExec.DurationMs = &dur

	if err != nil {
		stepExec.Status = "failed"
		stepExec.Error = sanitizeError(err.Error())
		if updateErr := e.persister.UpdateStepExecution(ctx, stepExec); updateErr != nil {
			fmt.Printf("failed to update step execution status to failed: %v\n", updateErr)
		}
		e.broadcastStepUpdate(runID, step, "failed", stepExec.Error, nil, retries)
		return err
	}

	stepExec.Status = "success"
	if result != nil {
		outputJSON, _ := json.Marshal(result)
		var outputMap domain.JSONB
		json.Unmarshal(outputJSON, &outputMap)
		stepExec.Output = outputMap
	}
	if updateErr := e.persister.UpdateStepExecution(ctx, stepExec); updateErr != nil {
		fmt.Printf("failed to update step execution status to success: %v\n", updateErr)
		return fmt.Errorf("failed to update step success: %w", updateErr)
	}

	outputs.Store(step.ID, result)

	if step.Type == "conditional" || step.Type == "condition" {
		if result != nil {
			if condResult, ok := result["condition_result"].(bool); ok && !condResult {
				skipped.Store(step.ID+"_condition", true)
			}
		}
	}

	e.broadcastStepUpdate(runID, step, "success", "", result, retries)
	return nil
}

func (e *Engine) shouldSkip(step domain.DAGStep, outputs *sync.Map, skipped *sync.Map) bool {
	for _, dep := range step.DependsOn {
		if _, isSkipped := skipped.Load(dep); isSkipped {
			return true
		}
		// skip if parent conditional evaluated to false
		if _, condFalse := skipped.Load(dep + "_condition"); condFalse {
			return true
		}
	}
	return false
}

func (e *Engine) collectInputs(step domain.DAGStep, outputs *sync.Map) map[string]interface{} {
	inputs := make(map[string]interface{})
	for _, dep := range step.DependsOn {
		if val, ok := outputs.Load(dep); ok {
			inputs[dep] = val
		}
	}
	return inputs
}

func (e *Engine) createStepRecord(runID uuid.UUID, step domain.DAGStep) *domain.StepExecution {
	now := time.Now()
	return &domain.StepExecution{
		ID:        uuid.New(),
		RunID:     runID,
		StepID:    step.ID,
		StepName:  step.Name,
		Status:    "running",
		StartedAt: &now,
	}
}

func (e *Engine) recordStep(ctx context.Context, runID uuid.UUID, step domain.DAGStep, status string, output map[string]interface{}, errMsg string, retries int) {
	now := time.Now()
	se := &domain.StepExecution{
		ID:          uuid.New(),
		RunID:       runID,
		StepID:      step.ID,
		StepName:    step.Name,
		Status:      status,
		StartedAt:   &now,
		CompletedAt: &now,
		RetryCount:  retries,
		Error:       errMsg,
	}
	dur := 0
	se.DurationMs = &dur
	_ = e.persister.CreateStepExecution(ctx, se)
	e.broadcastStepUpdate(runID, step, status, errMsg, output, retries)
}

func (e *Engine) failRun(ctx context.Context, runID uuid.UUID, errMsg string) {
	_ = e.persister.UpdateStatus(ctx, runID, "failed", sanitizeError(errMsg))
	e.broadcastRunUpdate(runID, "failed", sanitizeError(errMsg))
}

func sanitizeError(errMsg string) string {
	if len(errMsg) > 1000 {
		return errMsg[:1000]
	}
	return errMsg
}

func (e *Engine) parseDefinition(def domain.JSONB) (*domain.DAGDefinition, error) {
	return domain.ParseDAGDefinition(def)
}

func (e *Engine) broadcastStepUpdate(runID uuid.UUID, step domain.DAGStep, status, errMsg string, output map[string]interface{}, retries int) {
	if e.broadcaster == nil {
		return
	}
	e.broadcaster.BroadcastStepUpdate(StepUpdateEvent{
		RunID:     runID,
		StepID:    step.ID,
		StepName:  step.Name,
		Status:    status,
		Error:     errMsg,
		Output:    output,
		Retry:     retries,
		Timestamp: time.Now(),
	})
}

func (e *Engine) broadcastRunUpdate(runID uuid.UUID, status, errMsg string) {
	if e.broadcaster == nil {
		return
	}
	e.broadcaster.BroadcastRunUpdate(RunUpdateEvent{
		RunID:     runID,
		Status:    status,
		Error:     errMsg,
		Timestamp: time.Now(),
	})
}
