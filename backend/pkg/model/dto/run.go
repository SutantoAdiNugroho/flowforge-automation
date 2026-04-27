package dto

import "flowforge-automation-backend/pkg/model/domain"

type TriggerRunRequest struct {
	TriggeredBy string                 `json:"triggered_by,omitempty"`
	Inputs      map[string]interface{} `json:"inputs,omitempty"`
}

type RunDetailResponse struct {
	Run   domain.WorkflowRun    `json:"run"`
	Steps []domain.StepExecution `json:"steps"`
}
