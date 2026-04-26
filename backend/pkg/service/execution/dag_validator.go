package execution

import (
	"encoding/json"
	"fmt"
	"strings"

	"flowforge-automation-backend/pkg/model/domain"
)

type DAGValidator struct{}

func NewDAGValidator() *DAGValidator {
	return &DAGValidator{}
}

func (v *DAGValidator) Validate(definition domain.JSONB) (*domain.DAGValidationResult, error) {
	jsonBytes, err := json.Marshal(definition)
	if err != nil {
		return &domain.DAGValidationResult{
			Valid: false,
			Errors: []domain.DAGValidationError{
				{
					Field:   "definition",
					Message: "Invalid JSON format: " + err.Error(),
				},
			},
		}, nil
	}

	var dag domain.DAGDefinition
	if err := json.Unmarshal(jsonBytes, &dag); err != nil {
		return &domain.DAGValidationResult{
			Valid: false,
			Errors: []domain.DAGValidationError{
				{
					Field:   "definition",
					Message: "Invalid JSON format: " + err.Error(),
				},
			},
		}, nil
	}

	result := &domain.DAGValidationResult{
		Valid:  true,
		Errors: []domain.DAGValidationError{},
	}

	if err := v.validateStructure(&dag, result); err != nil {
		result.Valid = false
		return result, err
	}

	v.validateSteps(&dag, result)
	v.validateDependencies(&dag, result)
	v.detectCycles(&dag, result)

	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result, nil
}

func (v *DAGValidator) validateStructure(dag *domain.DAGDefinition, result *domain.DAGValidationResult) error {
	if dag == nil {
		result.Errors = append(result.Errors, domain.DAGValidationError{
			Field:   "root",
			Message: "DAG definition cannot be null",
		})
		return nil
	}

	if len(dag.Steps) == 0 {
		result.Errors = append(result.Errors, domain.DAGValidationError{
			Field:   "steps",
			Message: "DAG must have at least one step",
		})
	}

	return nil
}

func (v *DAGValidator) validateSteps(dag *domain.DAGDefinition, result *domain.DAGValidationResult) {
	seenIDs := make(map[string]bool)
	validTypes := map[string]bool{
		"http":      true,
		"script":    true,
		"condition": true,
		"parallel":  true,
		"delay":     true,
	}

	for i, step := range dag.Steps {
		if strings.TrimSpace(step.ID) == "" {
			result.Errors = append(result.Errors, domain.DAGValidationError{
				StepID:  fmt.Sprintf("step_%d", i),
				Field:   "id",
				Message: "Step ID cannot be empty",
			})
			continue
		}

		if strings.TrimSpace(step.Name) == "" {
			result.Errors = append(result.Errors, domain.DAGValidationError{
				StepID:  step.ID,
				Field:   "name",
				Message: "Step name cannot be empty",
			})
		}

		if seenIDs[step.ID] {
			result.Errors = append(result.Errors, domain.DAGValidationError{
				StepID:  step.ID,
				Field:   "id",
				Message: "Duplicate step ID",
			})
		}
		seenIDs[step.ID] = true

		if !validTypes[step.Type] {
			result.Errors = append(result.Errors, domain.DAGValidationError{
				StepID:  step.ID,
				Field:   "type",
				Message: fmt.Sprintf("Invalid step type: %s", step.Type),
			})
		}

		if step.Timeout != 0 && step.Timeout < 1000 {
			result.Errors = append(result.Errors, domain.DAGValidationError{
				StepID:  step.ID,
				Field:   "timeout",
				Message: "Timeout must be at least 1000ms",
			})
		}

		if step.RetryPolicy != nil {
			if step.RetryPolicy.MaxRetries < 0 {
				result.Errors = append(result.Errors, domain.DAGValidationError{
					StepID:  step.ID,
					Field:   "retry_policy.max_retries",
					Message: "Max retries cannot be negative",
				})
			}
			if step.RetryPolicy.InitialDelayMs < 100 {
				result.Errors = append(result.Errors, domain.DAGValidationError{
					StepID:  step.ID,
					Field:   "retry_policy.initial_delay_ms",
					Message: "Initial delay must be at least 100ms",
				})
			}
		}
	}
}

func (v *DAGValidator) validateDependencies(dag *domain.DAGDefinition, result *domain.DAGValidationResult) {
	stepIDs := make(map[string]bool)
	for _, step := range dag.Steps {
		stepIDs[step.ID] = true
	}

	for _, step := range dag.Steps {
		for _, depID := range step.DependsOn {
			if !stepIDs[depID] {
				result.Errors = append(result.Errors, domain.DAGValidationError{
					StepID:  step.ID,
					Field:   "depends_on",
					Message: fmt.Sprintf("Referenced step '%s' not found", depID),
				})
			}

			if depID == step.ID {
				result.Errors = append(result.Errors, domain.DAGValidationError{
					StepID:  step.ID,
					Field:   "depends_on",
					Message: "Step cannot depend on itself",
				})
			}
		}
	}
}

func (v *DAGValidator) detectCycles(dag *domain.DAGDefinition, result *domain.DAGValidationResult) {
	graph := make(map[string][]string)
	for _, step := range dag.Steps {
		graph[step.ID] = step.DependsOn
	}

	visited := make(map[string]int)
	for stepID := range graph {
		visited[stepID] = 0
	}

	for stepID := range graph {
		if visited[stepID] == 0 {
			if v.hasCycle(stepID, graph, visited) {
				result.Errors = append(result.Errors, domain.DAGValidationError{
					Field:   "dependencies",
					Message: fmt.Sprintf("Circular dependency detected involving step '%s'", stepID),
				})
			}
		}
	}
}

func (v *DAGValidator) hasCycle(node string, graph map[string][]string, visited map[string]int) bool {
	visited[node] = 1

	for _, neighbor := range graph[node] {
		if visited[neighbor] == 1 {
			return true
		}
		if visited[neighbor] == 0 && v.hasCycle(neighbor, graph, visited) {
			return true
		}
	}

	visited[node] = 2
	return false
}
