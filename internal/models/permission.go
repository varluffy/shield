package models

import (
	"gorm.io/gorm"
)

// Permission 权限模型
type Permission struct {
	BaseModel
	Code         string `gorm:"type:varchar(100);not null;uniqueIndex" json:"code"`
	Name         string `gorm:"type:varchar(100);not null" json:"name"`
	Description  string `gorm:"type:text" json:"description"`
	Type         string `gorm:"type:varchar(20);not null" json:"type"`      // menu, button, api
	Scope        string `gorm:"type:varchar(20);not null" json:"scope"`     // system, tenant
	ParentCode   string `gorm:"type:varchar(100);index" json:"parent_code"` // 父权限编码
	ResourcePath string `gorm:"type:varchar(200)" json:"resource_path"`     // API路径
	Method       string `gorm:"type:varchar(10)" json:"method"`             // HTTP方法
	SortOrder    int    `gorm:"default:0" json:"sort_order"`                // 排序
	IsBuiltin    bool   `gorm:"default:false" json:"is_builtin"`            // 是否内置
	IsActive     bool   `gorm:"default:true" json:"is_active"`              // 是否启用
	Module       string `gorm:"type:varchar(50)" json:"module"`             // 所属模块
	
	// 菜单专用字段
	MenuIcon      string `gorm:"type:varchar(100)" json:"menu_icon"`       // 菜单图标
	MenuPath      string `gorm:"type:varchar(200)" json:"menu_path"`       // 菜单路径
	MenuComponent string `gorm:"type:varchar(200)" json:"menu_component"`  // 前端组件路径
	MenuVisible   bool   `gorm:"default:true" json:"menu_visible"`         // 菜单是否可见
}

func (Permission) TableName() string {
	return "permissions"
}

// BeforeCreate 创建前钩子
func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.UUID == "" {
		p.UUID = GenerateUUID()
	}
	return nil
}

// FieldPermission 字段权限配置表
type FieldPermission struct {
	BaseModel
	EntityTable  string `gorm:"type:varchar(100);not null;index" json:"entity_table"`    // 表名
	FieldName    string `gorm:"type:varchar(100);not null;index" json:"field_name"`      // 字段名
	FieldLabel   string `gorm:"type:varchar(100);not null" json:"field_label"`           // 字段显示名
	FieldType    string `gorm:"type:varchar(50);not null" json:"field_type"`             // 字段类型
	DefaultValue string `gorm:"type:varchar(20);default:'default'" json:"default_value"` // 默认权限值
	Description  string `gorm:"type:text" json:"description"`                            // 字段描述
	SortOrder    int    `gorm:"default:0" json:"sort_order"`                             // 排序
	IsActive     bool   `gorm:"default:true" json:"is_active"`                           // 是否启用
}

func (FieldPermission) TableName() string {
	return "field_permissions"
}

// BeforeCreate 创建前钩子
func (fp *FieldPermission) BeforeCreate(tx *gorm.DB) error {
	if fp.UUID == "" {
		fp.UUID = GenerateUUID()
	}
	return nil
}

// RoleFieldPermission 角色字段权限表
type RoleFieldPermission struct {
	BaseModelWithoutUUID
	TenantID       uint64 `gorm:"not null;index" json:"tenant_id"`
	RoleID         uint64 `gorm:"not null;index" json:"role_id"`
	EntityTable    string `gorm:"type:varchar(100);not null;index" json:"entity_table"`
	FieldName      string `gorm:"type:varchar(100);not null;index" json:"field_name"`
	PermissionType string `gorm:"type:varchar(20);not null" json:"permission_type"` // default, hidden, readonly
}

func (RoleFieldPermission) TableName() string {
	return "role_field_permissions"
}