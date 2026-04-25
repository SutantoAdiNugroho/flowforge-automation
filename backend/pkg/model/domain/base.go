package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"type:timestamptz;default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamptz;default:current_timestamp" json:"updated_at"`
	DeletedAt *time.Time `gorm:"type:timestamptz;index" json:"deleted_at,omitempty"`
	UpdatedBy *uuid.UUID `gorm:"type:uuid" json:"updated_by,omitempty"`
}

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte("{}"), &j)
	}
	return json.Unmarshal(bytes, &j)
}
