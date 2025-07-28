package models

import (
	"gorm.io/gorm"
)

// Tenant 租户模型
type Tenant struct {
	BaseModel
	Name       string `gorm:"type:varchar(100);not null" json:"name"`
	Domain     string `gorm:"type:varchar(100);uniqueIndex" json:"domain"`
	Status     string `gorm:"type:varchar(20);default:'active'" json:"status"` // active, inactive, suspended
	Plan       string `gorm:"type:varchar(50);default:'basic'" json:"plan"`
	MaxUsers   int    `gorm:"default:100" json:"max_users"`
	MaxStorage int64  `gorm:"default:1073741824" json:"max_storage"`
	Settings   string `gorm:"type:json" json:"settings"`
}

func (Tenant) TableName() string {
	return "tenants"
}

// BeforeCreate 创建前钩子
func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.UUID == "" {
		t.UUID = GenerateUUID()
	}
	return nil
}