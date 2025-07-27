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
		t.UUID = uuid.New().String()
	}
	return nil
}

// User 用户模型
type User struct {
	TenantModel
	Email                string     `gorm:"type:varchar(255);not null;uniqueIndex:uk_tenant_email" json:"email"`
	Password             string     `gorm:"type:varchar(255);not null" json:"-"`
	Name                 string     `gorm:"type:varchar(100)" json:"name"`
	Avatar               string     `gorm:"type:varchar(500)" json:"avatar"`
	Phone                string     `gorm:"type:varchar(20)" json:"phone"`
	Status               string     `gorm:"type:varchar(20);default:'active'" json:"status"` // active, inactive, locked
	EmailVerifiedAt      *time.Time `json:"email_verified_at"`
	LastLoginAt          *time.Time `json:"last_login_at"`
	LoginCount           int        `gorm:"default:0" json:"login_count"`
	FailedLoginAttempts  int        `gorm:"default:0" json:"failed_login_attempts"`
	LockedUntil          *time.Time `json:"locked_until"`
	Timezone             string     `gorm:"type:varchar(50);default:'Asia/Shanghai'" json:"timezone"`
	Language             string     `gorm:"type:varchar(10);default:'zh'" json:"language"`
}

func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UUID == "" {
		u.UUID = uuid.New().String()
	}
	// 从上下文获取租户ID
	if u.TenantID == 0 {
		if tenantID := tx.Statement.Context.Value("tenant_id"); tenantID != nil {
			if tid, ok := tenantID.(uint64); ok {
				u.TenantID = tid
			}
		}
	}
	return nil
}

// Permission 权限模型
type Permission struct {
	BaseModel
	Code         string `gorm:"type:varchar(100);not null;uniqueIndex" json:"code"`
	Name         string `gorm:"type:varchar(100);not null" json:"name"`
	Description  string `gorm:"type:text" json:"description"`
	Type         string `gorm:"type:varchar(20);not null" json:"type"`        // menu, button, api
	Scope        string `gorm:"type:varchar(20);not null" json:"scope"`       // system, tenant
	ParentCode   string `gorm:"type:varchar(100);index" json:"parent_code"`   // 父权限编码
	ResourcePath string `gorm:"type:varchar(200)" json:"resource_path"`       // API路径
	Method       string `gorm:"type:varchar(10)" json:"method"`               // HTTP方法
	SortOrder    int    `gorm:"default:0" json:"sort_order"`                 // 排序
	IsBuiltin    bool   `gorm:"default:false" json:"is_builtin"`             // 是否内置
	IsActive     bool   `gorm:"default:true" json:"is_active"`               // 是否启用
	Module       string `gorm:"type:varchar(50)" json:"module"`              // 所属模块
}

func (Permission) TableName() string {
	return "permissions"
}

// BeforeCreate 创建前钩子
func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.UUID == "" {
		p.UUID = uuid.New().String()
	}
	return nil
}

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
		r.UUID = uuid.New().String()
	}
	// 从上下文获取租户ID
	if r.TenantID == 0 {
		if tenantID := tx.Statement.Context.Value("tenant_id"); tenantID != nil {
			if tid, ok := tenantID.(uint64); ok {
				r.TenantID = tid
			}
		}
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

// RefreshToken 刷新令牌模型（不需要UUID）
type RefreshToken struct {
	BaseModelWithoutUUID
	UserID    uint64    `gorm:"not null;index" json:"user_id"`
	TenantID  uint64    `gorm:"not null;index" json:"tenant_id"`
	TokenHash string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"token_hash"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IsRevoked bool      `gorm:"default:false" json:"is_revoked"`
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	IPAddress string    `gorm:"type:varchar(45)" json:"ip_address"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// LoginAttempt 登录尝试记录模型（不需要UUID）
type LoginAttempt struct {
	BaseModelWithoutUUID
	UserID        uint64    `gorm:"index" json:"user_id"`
	TenantID      uint64    `gorm:"index" json:"tenant_id"`
	Email         string    `gorm:"type:varchar(255);index" json:"email"`
	IPAddress     string    `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent     string    `gorm:"type:text" json:"user_agent"`
	Success       bool      `gorm:"default:false" json:"success"`
	FailureReason string    `gorm:"type:varchar(100)" json:"failure_reason"`
}

func (LoginAttempt) TableName() string {
	return "login_attempts"
}

// UserProfile 用户资料模型（保留UUID，可能对外暴露）
type UserProfile struct {
	BaseModel
	UserID    uint64     `gorm:"not null;uniqueIndex" json:"user_id"`
	FirstName string     `gorm:"type:varchar(50)" json:"first_name"`
	LastName  string     `gorm:"type:varchar(50)" json:"last_name"`
	Bio       string     `gorm:"type:text" json:"bio"`
	Birthday  *time.Time `json:"birthday"`
	Gender    string     `gorm:"type:varchar(10)" json:"gender"`
	Address   string     `gorm:"type:text" json:"address"`
}

func (UserProfile) TableName() string {
	return "user_profiles"
}

// BeforeCreate 创建前钩子
func (up *UserProfile) BeforeCreate(tx *gorm.DB) error {
	if up.UUID == "" {
		up.UUID = uuid.New().String()
	}
	return nil
}

// FieldPermission 字段权限配置表
type FieldPermission struct {
	BaseModel
	EntityTable  string `gorm:"type:varchar(100);not null;index" json:"entity_table"`   // 表名
	FieldName    string `gorm:"type:varchar(100);not null;index" json:"field_name"`     // 字段名
	FieldLabel   string `gorm:"type:varchar(100);not null" json:"field_label"`          // 字段显示名
	FieldType    string `gorm:"type:varchar(50);not null" json:"field_type"`            // 字段类型
	DefaultValue string `gorm:"type:varchar(20);default:'default'" json:"default_value"` // 默认权限值
	Description  string `gorm:"type:text" json:"description"`                           // 字段描述
	SortOrder    int    `gorm:"default:0" json:"sort_order"`                           // 排序
	IsActive     bool   `gorm:"default:true" json:"is_active"`                         // 是否启用
}

func (FieldPermission) TableName() string {
	return "field_permissions"
}

// BeforeCreate 创建前钩子
func (fp *FieldPermission) BeforeCreate(tx *gorm.DB) error {
	if fp.UUID == "" {
		fp.UUID = uuid.New().String()
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

// PermissionAuditLog 权限操作审计日志
type PermissionAuditLog struct {
	BaseModel
	TenantID     uint64 `gorm:"not null;index" json:"tenant_id"`
	OperatorID   uint64 `gorm:"not null;index" json:"operator_id"`       // 操作人
	TargetType   string `gorm:"type:varchar(50);not null" json:"target_type"` // user, role, permission
	TargetID     uint64 `gorm:"not null;index" json:"target_id"`         // 目标ID
	Action         string `gorm:"type:varchar(50);not null" json:"action"`         // grant, revoke, create, delete
	PermissionCode string `gorm:"type:varchar(100)" json:"permission_code"`        // 权限代码
	OldValue       string `gorm:"type:text" json:"old_value"`                      // 变更前值
	NewValue     string `gorm:"type:text" json:"new_value"`              // 变更后值
	Reason       string `gorm:"type:text" json:"reason"`                 // 操作原因
	IPAddress    string `gorm:"type:varchar(45)" json:"ip_address"`      // 操作IP
	UserAgent    string `gorm:"type:text" json:"user_agent"`             // 用户代理
}

func (PermissionAuditLog) TableName() string {
	return "permission_audit_logs"
}

// BeforeCreate 创建前钩子
func (pal *PermissionAuditLog) BeforeCreate(tx *gorm.DB) error {
	if pal.UUID == "" {
		pal.UUID = uuid.New().String()
	}
	return nil
}

// PhoneBlacklist 手机号黑名单模型
type PhoneBlacklist struct {
	TenantModel
	PhoneMD5    string `gorm:"type:char(32);not null;uniqueIndex:uk_tenant_phone_md5" json:"phone_md5"`
	Source      string `gorm:"type:varchar(50);not null" json:"source"`                 // manual, import, api
	Reason      string `gorm:"type:varchar(200)" json:"reason"`                         // 加入黑名单原因
	OperatorID  uint64 `gorm:"index" json:"operator_id"`                               // 操作人ID
	IsActive    bool   `gorm:"default:true" json:"is_active"`                          // 是否有效
}

func (PhoneBlacklist) TableName() string {
	return "phone_blacklists"
}

// BeforeCreate 创建前钩子
func (pb *PhoneBlacklist) BeforeCreate(tx *gorm.DB) error {
	if pb.UUID == "" {
		pb.UUID = uuid.New().String()
	}
	// 从上下文获取租户ID
	if pb.TenantID == 0 {
		if tenantID := tx.Statement.Context.Value("tenant_id"); tenantID != nil {
			if tid, ok := tenantID.(uint64); ok {
				pb.TenantID = tid
			}
		}
	}
	return nil
}

// BlacklistApiCredential 黑名单API密钥模型
type BlacklistApiCredential struct {
	TenantModel
	APIKey      string `gorm:"type:varchar(64);not null;uniqueIndex" json:"api_key"`
	APISecret   string `gorm:"type:varchar(128);not null" json:"api_secret"`
	Name        string `gorm:"type:varchar(100);not null" json:"name"`                // 密钥名称
	Description string `gorm:"type:text" json:"description"`                         // 描述
	RateLimit   int    `gorm:"default:1000" json:"rate_limit"`                       // 每秒请求限制
	Status      string `gorm:"type:varchar(20);default:'active'" json:"status"`      // active, inactive, suspended
	LastUsedAt  *time.Time `json:"last_used_at"`                                     // 最后使用时间
	ExpiresAt   *time.Time `json:"expires_at"`                                       // 过期时间
}

func (BlacklistApiCredential) TableName() string {
	return "blacklist_api_credentials"
}

// BeforeCreate 创建前钩子
func (bac *BlacklistApiCredential) BeforeCreate(tx *gorm.DB) error {
	if bac.UUID == "" {
		bac.UUID = uuid.New().String()
	}
	// 从上下文获取租户ID
	if bac.TenantID == 0 {
		if tenantID := tx.Statement.Context.Value("tenant_id"); tenantID != nil {
			if tid, ok := tenantID.(uint64); ok {
				bac.TenantID = tid
			}
		}
	}
	return nil
}

// BlacklistQueryLog 黑名单查询日志模型（用于统计分析）
type BlacklistQueryLog struct {
	BaseModelWithoutUUID
	TenantID      uint64    `gorm:"not null;index" json:"tenant_id"`
	APIKey        string    `gorm:"type:varchar(64);not null;index" json:"api_key"`
	PhoneMD5      string    `gorm:"type:char(32);not null" json:"phone_md5"`
	IsHit         bool      `gorm:"default:false" json:"is_hit"`                    // 是否命中黑名单
	ResponseTime  int       `gorm:"not null" json:"response_time"`                 // 响应时间(毫秒)
	ClientIP      string    `gorm:"type:varchar(45)" json:"client_ip"`             // 客户端IP
	UserAgent     string    `gorm:"type:text" json:"user_agent"`                   // 用户代理
	RequestID     string    `gorm:"type:varchar(64);index" json:"request_id"`      // 请求ID
}

func (BlacklistQueryLog) TableName() string {
	return "blacklist_query_logs"
}
