package enum

// WorkflowRunStatus represents status values for workflow runs.
type WorkflowRunStatus string

const (
	WorkflowRunStatusPending   WorkflowRunStatus = "pending"
	WorkflowRunStatusRunning   WorkflowRunStatus = "running"
	WorkflowRunStatusSuccess   WorkflowRunStatus = "success"
	WorkflowRunStatusFailed    WorkflowRunStatus = "failed"
	WorkflowRunStatusTimeout   WorkflowRunStatus = "timeout"
	WorkflowRunStatusCancelled WorkflowRunStatus = "cancelled"
)
