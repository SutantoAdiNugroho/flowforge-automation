package domain

type Tenant struct {
	BaseModel
	Name string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Slug string `gorm:"column:slug;type:varchar(100);uniqueIndex;not null" json:"slug"`
}

func (Tenant) TableName() string {
	return "tenants"
}
