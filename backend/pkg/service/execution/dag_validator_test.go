package execution

import (
	"encoding/json"
	"testing"

	"flowforge-automation-backend/pkg/model/domain"
	"github.com/stretchr/testify/assert"
)

func newJSONB(jsonStr string) domain.JSONB {
	var data domain.JSONB
	json.Unmarshal([]byte(jsonStr), &data)
	return data
}

func TestDAGValidatorValidDefinition(t *testing.T) {
	validator := NewDAGValidator()
	definition := newJSONB(`{
		"version": "1.0",
		"steps": [
			{
				"id": "step1",
				"name": "First Step",
				"type": "http",
				"config": {"url": "http://example.com"},
				"timeout": 5000
			},
			{
				"id": "step2",
				"name": "Second Step",
				"type": "script",
				"depends_on": ["step1"]
			}
		]
	}`)

	result, err := validator.Validate(definition)
	assert.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestDAGValidatorDuplicateStepID(t *testing.T) {
	validator := NewDAGValidator()
	definition := newJSONB(`{
		"version": "1.0",
		"steps": [
			{
				"id": "step1",
				"name": "First Step",
				"type": "http"
			},
			{
				"id": "step1",
				"name": "Duplicate Step",
				"type": "script"
			}
		]
	}`)

	result, err := validator.Validate(definition)
	assert.NoError(t, err)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
	assert.Equal(t, "Duplicate step ID", result.Errors[0].Message)
}

func TestDAGValidatorInvalidStepType(t *testing.T) {
	validator := NewDAGValidator()
	definition := newJSONB(`{
		"version": "1.0",
		"steps": [
			{
				"id": "step1",
				"name": "First Step",
				"type": "invalid_type"
			}
		]
	}`)

	result, err := validator.Validate(definition)
	assert.NoError(t, err)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
}

func TestDAGValidatorMissingDependency(t *testing.T) {
	validator := NewDAGValidator()
	definition := newJSONB(`{
		"version": "1.0",
		"steps": [
			{
				"id": "step1",
				"name": "First Step",
				"type": "http",
				"depends_on": ["nonexistent"]
			}
		]
	}`)

	result, err := validator.Validate(definition)
	assert.NoError(t, err)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
	assert.Contains(t, result.Errors[0].Message, "not found")
}

func TestDAGValidatorCircularDependency(t *testing.T) {
	validator := NewDAGValidator()
	definition := newJSONB(`{
		"version": "1.0",
		"steps": [
			{
				"id": "step1",
				"name": "First Step",
				"type": "http",
				"depends_on": ["step2"]
			},
			{
				"id": "step2",
				"name": "Second Step",
				"type": "script",
				"depends_on": ["step1"]
			}
		]
	}`)

	result, err := validator.Validate(definition)
	assert.NoError(t, err)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
}

func TestDAGValidatorSelfDependency(t *testing.T) {
	validator := NewDAGValidator()
	definition := newJSONB(`{
		"version": "1.0",
		"steps": [
			{
				"id": "step1",
				"name": "First Step",
				"type": "http",
				"depends_on": ["step1"]
			}
		]
	}`)

	result, err := validator.Validate(definition)
	assert.NoError(t, err)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
}

func TestDAGValidatorEmptySteps(t *testing.T) {
	validator := NewDAGValidator()
	definition := newJSONB(`{
		"version": "1.0",
		"steps": []
	}`)

	result, err := validator.Validate(definition)
	assert.NoError(t, err)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
}

func TestDAGValidatorInvalidJSON(t *testing.T) {
	validator := NewDAGValidator()
	definition := newJSONB(`{invalid json`)

	result, err := validator.Validate(definition)
	assert.NoError(t, err)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
}

func TestDAGValidatorRetryPolicy(t *testing.T) {
	validator := NewDAGValidator()
	definition := newJSONB(`{
		"version": "1.0",
		"steps": [
			{
				"id": "step1",
				"name": "First Step",
				"type": "http",
				"retry_policy": {
					"max_retries": 3,
					"initial_delay_ms": 100,
					"max_delay_ms": 5000,
					"backoff_multiplier": 2.0
				}
			}
		]
	}`)

	result, err := validator.Validate(definition)
	assert.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}

func TestDAGValidatorInvalidRetryPolicy(t *testing.T) {
	validator := NewDAGValidator()
	definition := newJSONB(`{
		"version": "1.0",
		"steps": [
			{
				"id": "step1",
				"name": "First Step",
				"type": "http",
				"retry_policy": {
					"max_retries": -1,
					"initial_delay_ms": 50
				}
			}
		]
	}`)

	result, err := validator.Validate(definition)
	assert.NoError(t, err)
	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Errors)
}

func TestDAGValidatorMultipleLevelDependencies(t *testing.T) {
	validator := NewDAGValidator()
	definition := newJSONB(`{
		"version": "1.0",
		"steps": [
			{
				"id": "step1",
				"name": "First Step",
				"type": "http"
			},
			{
				"id": "step2",
				"name": "Second Step",
				"type": "script",
				"depends_on": ["step1"]
			},
			{
				"id": "step3",
				"name": "Third Step",
				"type": "condition",
				"depends_on": ["step2"]
			},
			{
				"id": "step4",
				"name": "Fourth Step",
				"type": "parallel",
				"depends_on": ["step1", "step3"]
			}
		]
	}`)

	result, err := validator.Validate(definition)
	assert.NoError(t, err)
	assert.True(t, result.Valid)
	assert.Empty(t, result.Errors)
}
