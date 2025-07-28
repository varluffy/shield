// Package models contains database models and entity definitions.
// It defines the data structures that map to database tables.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel 基础模型（带UUID）
type BaseModel struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID      string         `gorm:"type:char(36);not null;uniqueIndex" json:"uuid"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BaseModelWithoutUUID 基础模型（不带UUID）
type BaseModelWithoutUUID struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TenantModel 租户模型（带UUID）
type TenantModel struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID      string         `gorm:"type:char(36);not null;uniqueIndex" json:"uuid"`
	TenantID  uint64         `gorm:"not null;index" json:"tenant_id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TenantModelWithoutUUID 租户模型（不带UUID）
type TenantModelWithoutUUID struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  uint64         `gorm:"not null;index" json:"tenant_id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// GenerateUUID 生成UUID的辅助函数
func GenerateUUID() string {
	return uuid.New().String()
}

// GetTenantIDFromContext 从上下文获取租户ID的辅助函数
func GetTenantIDFromContext(tx *gorm.DB) uint64 {
	if tenantID := tx.Statement.Context.Value("tenant_id"); tenantID != nil {
		if tid, ok := tenantID.(uint64); ok {
			return tid
		}
	}
	return 0
}