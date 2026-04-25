package domain

import (
	"time"

	"flowforge-automation-backend/pkg/model/domain/enum"
	"github.com/google/uuid"
)

type StepExecution struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RunID      uuid.UUID `gorm:"column:run_id;type:uuid;index;not null" json:"run_id"`
	Run        *WorkflowRun `gorm:"foreignKey:RunID;references:ID;constraint:OnDelete:CASCADE" json:"run,omitempty"`
	StepID     string    `gorm:"column:step_id;type:varchar(100);index;not null" json:"step_id"`
	StepName   string    `gorm:"column:step_name;type:varchar(255);not null" json:"step_name"`
	Status     string    `gorm:"column:status;type:varchar(20);index;not null" json:"status"` // pending, running, success, failed, skipped, timeout
	StartedAt  *time.Time `gorm:"column:started_at;type:timestamptz" json:"started_at,omitempty"`
	CompletedAt *time.Time `gorm:"column:completed_at;type:timestamptz" json:"completed_at,omitempty"`
	DurationMs *int      `gorm:"column:duration_ms;type:integer" json:"duration_ms,omitempty"`
	RetryCount int       `gorm:"column:retry_count;type:integer;default:0" json:"retry_count"`
	Output     JSONB     `gorm:"column:output;type:jsonb" json:"output,omitempty"`
	Error      string    `gorm:"column:error;type:text" json:"error,omitempty"`
	CreatedAt  time.Time `gorm:"column:created_at;type:timestamptz;default:current_timestamp" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;type:timestamptz;default:current_timestamp" json:"updated_at"`
}

func (StepExecution) TableName() string {
	return "step_executions"
}

func (s *StepExecution) IsSuccess() bool {
	return s.Status == string(enum.StepExecutionStatusSuccess)
}

func (s *StepExecution) IsFailed() bool {
	return s.Status == string(enum.StepExecutionStatusFailed)
}

func (s *StepExecution) IsRunning() bool {
	return s.Status == string(enum.StepExecutionStatusRunning)
}
