package dto

import "flowforge-automation-backend/pkg/model/domain"

type CreateWorkflowRequest struct {
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	Definition     domain.JSONB `json:"definition"`
	TriggerType    string       `json:"trigger_type"`
	CronExpression string       `json:"cron_expression"`
	IsActive       *bool        `json:"is_active,omitempty"`
}

type UpdateWorkflowRequest struct {
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	Definition     domain.JSONB `json:"definition"`
	TriggerType    string       `json:"trigger_type"`
	CronExpression string       `json:"cron_expression"`
	IsActive       *bool        `json:"is_active,omitempty"`
}

type CreateWorkflowVersionRequest struct {
	ImportFromVersion *int         `json:"import_from_version,omitempty"`
	Name              string       `json:"name"`
	Description       string       `json:"description"`
	Definition        domain.JSONB `json:"definition"`
	TriggerType       string       `json:"trigger_type"`
	CronExpression    string       `json:"cron_expression"`
}
