package domain

import (
	"flowforge-automation-backend/pkg/model/domain/enum"
	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	BaseModel
	TenantID     uuid.UUID `gorm:"column:tenant_id;type:uuid;uniqueIndex:idx_tenant_email,priority:1;not null" json:"tenant_id"`
	Tenant       *Tenant   `gorm:"foreignKey:TenantID;references:ID;constraint:OnDelete:CASCADE" json:"tenant,omitempty"`
	Email        string    `gorm:"column:email;type:varchar(255);uniqueIndex:idx_tenant_email,priority:2;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	Role         string    `gorm:"column:role;type:varchar(20);default:'viewer';not null" json:"role"` // admin, editor, viewer
	IsActive     bool      `gorm:"column:is_active;type:boolean;default:true;not null" json:"is_active"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) IsAdmin() bool {
	return u.Role == string(enum.UserRoleAdmin)
}

func (u *User) IsEditor() bool {
	return u.Role == string(enum.UserRoleEditor)
}

func (u *User) IsViewer() bool {
	return u.Role == string(enum.UserRoleViewer)
}
