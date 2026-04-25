package domain

import "github.com/google/uuid"

type WorkflowVersion struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WorkflowID  uuid.UUID `gorm:"column:workflow_id;type:uuid;index;not null" json:"workflow_id"`
	Workflow    *Workflow `gorm:"foreignKey:WorkflowID;references:ID;constraint:OnDelete:CASCADE" json:"workflow,omitempty"`
	Version     int       `gorm:"column:version;type:integer;not null" json:"version"`
	Definition  JSONB     `gorm:"column:definition;type:jsonb;not null" json:"definition"`
	CreatedByID uuid.UUID `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	CreatedByUser *User   `gorm:"foreignKey:CreatedByID;references:ID" json:"created_by_user,omitempty"`
	CreatedAt   int64     `gorm:"column:created_at;type:timestamptz;default:current_timestamp" json:"created_at"`
}

func (WorkflowVersion) TableName() string {
	return "workflow_versions"
}
