package models

import (
	"time"

	"gorm.io/gorm"
)

// Role 角色模型
type Role struct {
	TenantModel
	Code        string `gorm:"type:varchar(100);not null;uniqueIndex:uk_tenant_code" json:"code"`
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Type        string `gorm:"type:varchar(20);default:'custom'" json:"type"` // system, custom
	IsActive    bool   `gorm:"default:true" json:"is_active"`
}

func (Role) TableName() string {
	return "roles"
}

// BeforeCreate 创建前钩子
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.UUID == "" {
		r.UUID = GenerateUUID()
	}
	// 从上下文获取租户ID
	if r.TenantID == 0 {
		r.TenantID = GetTenantIDFromContext(tx)
	}
	return nil
}

// UserRole 用户角色关联模型（不需要UUID）
type UserRole struct {
	BaseModelWithoutUUID
	UserID    uint64     `gorm:"not null;uniqueIndex:uk_user_role" json:"user_id"`
	RoleID    uint64     `gorm:"not null;uniqueIndex:uk_user_role" json:"role_id"`
	TenantID  uint64     `gorm:"not null;index" json:"tenant_id"`
	GrantedBy uint64     `gorm:"index" json:"granted_by"`
	GrantedAt time.Time  `gorm:"autoCreateTime" json:"granted_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	IsActive  bool       `gorm:"default:true" json:"is_active"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

// RolePermission 角色权限关联模型（不需要UUID）
type RolePermission struct {
	BaseModelWithoutUUID
	RoleID       uint64    `gorm:"not null;uniqueIndex:uk_role_permission" json:"role_id"`
	PermissionID uint64    `gorm:"not null;uniqueIndex:uk_role_permission" json:"permission_id"`
	GrantedBy    uint64    `gorm:"index" json:"granted_by"`
	GrantedAt    time.Time `gorm:"autoCreateTime" json:"granted_at"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}