// Package models contains database models and entity definitions.
package models

// Permission types
const (
	PermissionTypeMenu   = "menu"
	PermissionTypeButton = "button"
	PermissionTypeAPI    = "api"
)

// Permission scopes
const (
	ScopeSystem = "system" // 系统权限：只有系统管理员可见
	ScopeTenant = "tenant" // 租户权限：租户内可见
)

// Role types
const (
	RoleTypeSystem = "system" // 系统内置角色
	RoleTypeCustom = "custom" // 自定义角色
)

// Special role codes
const (
	RoleSystemAdmin = "system_admin" // 系统管理员
	RoleTenantAdmin = "tenant_admin" // 租户管理员
)

// Field permission types
const (
	FieldPermissionDefault  = "default"  // 默认：正常显示和编辑
	FieldPermissionHidden   = "hidden"   // 隐藏：不显示该字段
	FieldPermissionReadonly = "readonly" // 只读：显示但不能编辑
)

// Permission modules
const (
	ModuleUser      = "user"      // 用户管理
	ModuleRole      = "role"      // 角色管理
	ModuleSystem    = "system"    // 系统管理
	ModuleTenant    = "tenant"    // 租户管理
	ModuleCandidate = "candidate" // 候选人管理
	ModuleCompany   = "company"   // 企业管理
)

// Audit log actions
const (
	AuditActionGrant  = "grant"  // 授予权限
	AuditActionRevoke = "revoke" // 撤销权限
	AuditActionCreate = "create" // 创建
	AuditActionUpdate = "update" // 更新
	AuditActionDelete = "delete" // 删除
)

// Audit log target types
const (
	AuditTargetUser       = "user"
	AuditTargetRole       = "role"
	AuditTargetPermission = "permission"
)

// User status
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusLocked   = "locked"
)

// Tenant status
const (
	TenantStatusActive    = "active"
	TenantStatusInactive  = "inactive"
	TenantStatusSuspended = "suspended"
) 