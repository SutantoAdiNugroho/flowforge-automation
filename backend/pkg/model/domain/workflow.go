package domain

import "github.com/google/uuid"

type Workflow struct {
	BaseModel
	TenantID       uuid.UUID `gorm:"column:tenant_id;type:uuid;index;not null" json:"tenant_id"`
	Tenant         *Tenant   `gorm:"foreignKey:TenantID;references:ID;constraint:OnDelete:CASCADE" json:"tenant,omitempty"`
	Name           string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description    string    `gorm:"column:description;type:text" json:"description,omitempty"`
	Definition     JSONB     `gorm:"column:definition;type:jsonb;not null" json:"definition"`
	Version        int       `gorm:"column:version;type:integer;default:1;not null" json:"version"`
	IsActive       bool      `gorm:"column:is_active;type:boolean;default:true;not null" json:"is_active"`
	TriggerType    string    `gorm:"column:trigger_type;type:varchar(20);not null" json:"trigger_type"` // manual, cron, webhook
	CronExpression string    `gorm:"column:cron_expression;type:varchar(100)" json:"cron_expression,omitempty"`
	CreatedByID    uuid.UUID `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	CreatedByUser  *User     `gorm:"foreignKey:CreatedByID;references:ID" json:"created_by_user,omitempty"`
}

func (Workflow) TableName() string {
	return "workflows"
}
