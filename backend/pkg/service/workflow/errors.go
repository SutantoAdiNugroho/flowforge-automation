package workflow

import "errors"

var (
	ErrInvalidCreateWorkflowRequest = errors.New("invalid create workflow request")
	ErrInvalidUpdateWorkflowRequest = errors.New("invalid update workflow request")
	ErrInvalidWorkflowDefinition    = errors.New("invalid workflow definition")
	ErrWorkflowNotFound            = errors.New("workflow not found")
)
