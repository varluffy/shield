package models

import (
	"gorm.io/gorm"
)

// PermissionAuditLog 权限操作审计日志
type PermissionAuditLog struct {
	BaseModel
	TenantID       uint64 `gorm:"not null;index" json:"tenant_id"`
	OperatorID     uint64 `gorm:"not null;index" json:"operator_id"`            // 操作人
	TargetType     string `gorm:"type:varchar(50);not null" json:"target_type"` // user, role, permission
	TargetID       uint64 `gorm:"not null;index" json:"target_id"`              // 目标ID
	Action         string `gorm:"type:varchar(50);not null" json:"action"`      // grant, revoke, create, delete
	PermissionCode string `gorm:"type:varchar(100)" json:"permission_code"`     // 权限代码
	OldValue       string `gorm:"type:text" json:"old_value"`                   // 变更前值
	NewValue       string `gorm:"type:text" json:"new_value"`                   // 变更后值
	Reason         string `gorm:"type:text" json:"reason"`                      // 操作原因
	IPAddress      string `gorm:"type:varchar(45)" json:"ip_address"`           // 操作IP
	UserAgent      string `gorm:"type:text" json:"user_agent"`                  // 用户代理
}

func (PermissionAuditLog) TableName() string {
	return "permission_audit_logs"
}

// BeforeCreate 创建前钩子
func (pal *PermissionAuditLog) BeforeCreate(tx *gorm.DB) error {
	if pal.UUID == "" {
		pal.UUID = GenerateUUID()
	}
	return nil
}

// LoginAttempt 登录尝试记录模型（不需要UUID）
type LoginAttempt struct {
	BaseModelWithoutUUID
	UserID        uint64 `gorm:"index" json:"user_id"`
	TenantID      uint64 `gorm:"index" json:"tenant_id"`
	Email         string `gorm:"type:varchar(255);index" json:"email"`
	IPAddress     string `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent     string `gorm:"type:text" json:"user_agent"`
	Success       bool   `gorm:"default:false" json:"success"`
	FailureReason string `gorm:"type:varchar(100)" json:"failure_reason"`
}

func (LoginAttempt) TableName() string {
	return "login_attempts"
}