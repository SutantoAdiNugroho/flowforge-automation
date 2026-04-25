package enum

// StepExecutionStatus represents status values for step executions.
type StepExecutionStatus string

const (
	StepExecutionStatusPending StepExecutionStatus = "pending"
	StepExecutionStatusRunning StepExecutionStatus = "running"
	StepExecutionStatusSuccess StepExecutionStatus = "success"
	StepExecutionStatusFailed  StepExecutionStatus = "failed"
	StepExecutionStatusSkipped StepExecutionStatus = "skipped"
	StepExecutionStatusTimeout StepExecutionStatus = "timeout"
)
