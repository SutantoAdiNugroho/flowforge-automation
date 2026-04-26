package domain

import (
	"time"

	"flowforge-automation-backend/pkg/model/domain/enum"
	"github.com/google/uuid"
)

type WorkflowRun struct {
	BaseModel
	WorkflowID    uuid.UUID `gorm:"column:workflow_id;type:uuid;index;not null" json:"workflow_id"`
	Workflow      *Workflow `gorm:"foreignKey:WorkflowID;references:ID;constraint:OnDelete:CASCADE" json:"workflow,omitempty"`
	TenantID      uuid.UUID `gorm:"column:tenant_id;type:uuid;index;not null" json:"tenant_id"`
	Tenant        *Tenant   `gorm:"foreignKey:TenantID;references:ID;constraint:OnDelete:CASCADE" json:"tenant,omitempty"`
	Status        string    `gorm:"column:status;type:varchar(20);default:'pending';index;not null" json:"status"` // pending, running, success, failed, timeout, cancelled
	TriggeredBy   string    `gorm:"column:triggered_by;type:varchar(50);not null" json:"triggered_by"`              // manual, cron, webhook
	StartedAt     *time.Time `gorm:"column:started_at;type:timestamptz" json:"started_at,omitempty"`
	CompletedAt   *time.Time `gorm:"column:completed_at;type:timestamptz" json:"completed_at,omitempty"`
	ErrorMessage  string    `gorm:"column:error_message;type:text" json:"error_message,omitempty"`
	Metadata      JSONB     `gorm:"column:metadata;type:jsonb" json:"metadata,omitempty"`
}

func (WorkflowRun) TableName() string {
	return "workflow_runs"
}

func (w *WorkflowRun) Duration() int {
	if w.StartedAt == nil || w.CompletedAt == nil {
		return 0
	}
	return int(w.CompletedAt.Sub(*w.StartedAt).Seconds())
}

func (w *WorkflowRun) IsSuccess() bool {
	return w.Status == string(enum.WorkflowRunStatusSuccess)
}

func (w *WorkflowRun) IsFailed() bool {
	return w.Status == string(enum.WorkflowRunStatusFailed)
}

func (w *WorkflowRun) IsRunning() bool {
	return w.Status == string(enum.WorkflowRunStatusRunning)
}
