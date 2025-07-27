# UltraFit 权限控制系统设计

## 🎯 设计理念

UltraFit权限控制系统基于RBAC（基于角色的访问控制）模型，采用了**极简化和灵活性**的设计理念：

### 核心设计原则

1. **唯一硬编码**: 系统中只有`tenant_id = 0`是硬编码的，用于标识系统租户
2. **灵活角色命名**: 所有角色名称都可以自由定义，没有硬编码限制
3. **scope驱动**: 通过`permissions.scope`字段区分权限类型（system/tenant）
4. **自动化初始化**: 新租户创建时，自动分配所有`scope='tenant'`的权限

### 权限分类逻辑

```
权限作用域（Scope）：
├── system - 系统级权限，只有系统管理员可见
│   ├── 租户管理（创建、删除租户）
│   ├── 系统配置（修改系统参数）
│   └── 权限管理（创建、修改权限）
└── tenant - 租户级权限，租户内可见
    ├── 用户管理（增删改查用户）
    ├── 角色管理（创建、分配角色）
    └── 字段权限（配置字段可见性）
```

### 新租户初始化流程

```sql
-- 1. 创建新租户
INSERT INTO tenants (name, domain) VALUES ('新租户', 'new-tenant.com');

-- 2. 查询所有租户权限
SELECT * FROM permissions WHERE scope = 'tenant';

-- 3. 创建租户管理员角色
INSERT INTO roles (tenant_id, code, name, type) 
VALUES (new_tenant_id, 'tenant_admin', '租户管理员', 'system');

-- 4. 自动关联所有租户权限
-- (通过代码逻辑实现)
```

## 🏗️ RBAC架构设计

### 1. 核心概念

```
用户(User) ←→ 角色(Role) ←→ 权限(Permission)
```

### 2. 权限类型层级

```
菜单权限 (menu)
├── 按钮权限 (button)
│   └── API权限 (api)
└── 字段权限 (field)
```

### 3. 组件关系

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    用户     │    │    角色     │    │    权限     │
│    User     │───▶│    Role     │───▶│ Permission  │
└─────────────┘    └─────────────┘    └─────────────┘
                                             │
                                             ▼
                                    ┌─────────────────┐
                                    │   字段权限配置   │
                                    │ FieldPermission │
                                    └─────────────────┘
```

## 🔐 权限模型设计

### 1. 权限模型（优化后）

```go
// Permission 权限模型 - 精简设计
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
```

### 2. 核心字段说明

| 字段 | 用途 | 是否必填 | 说明 |
|------|------|---------|------|
| `code` | 权限唯一标识 | ✅ | 如：user_list_api |
| `scope` | 权限作用域 | ✅ | system/tenant，**核心分类字段** |
| `type` | 权限类型 | ✅ | menu/button/api |
| `module` | 权限模块 | ❌ | user/role/system等 |
| `parent_code` | 父权限 | ❌ | 构建权限树结构 |
| `resource_path` | API路径 | ❌ | 用于API权限验证 |
| `method` | HTTP方法 | ❌ | GET/POST/PUT/DELETE |
| `is_builtin` | 内置权限 | ❌ | 防止误删系统权限 |

### 3. 已移除的冗余字段

在最新优化中，我们移除了以下未使用的字段：
- `category` - 权限分类（未使用）
- `resource` - 资源标识（未使用）  
- `action` - 操作类型（未使用）
- `is_system` - 系统标识（与scope字段重复）

### 4. 权限常量定义

```go
// 权限类型常量
const (
    PermissionTypeMenu   = "menu"
    PermissionTypeButton = "button"
    PermissionTypeAPI    = "api"
)

// 权限作用域常量
const (
    ScopeSystem = "system" // 系统权限：只有系统管理员可见
    ScopeTenant = "tenant" // 租户权限：租户内可见
)
```

## 👑 角色模型设计

```go
// Role 角色模型
type Role struct {
    TenantModel
    Code        string `gorm:"type:varchar(100);not null;uniqueIndex:uk_tenant_code" json:"code"`
    Name        string `gorm:"type:varchar(100);not null" json:"name"`
    Description string `gorm:"type:text" json:"description"`
    Type        string `gorm:"type:varchar(20);default:'custom'" json:"type"` // system, custom
    IsActive    bool   `gorm:"default:true" json:"is_active"`
}

// 角色类型常量
const (
    RoleTypeSystem = "system" // 系统内置角色
    RoleTypeCustom = "custom" // 自定义角色
)

// 特殊角色代码
const (
    RoleSystemAdmin = "system_admin" // 系统管理员
    RoleTenantAdmin = "tenant_admin" // 租户管理员
)
```

### 3. 用户角色关联（实际实现）

```go
// UserRole 用户角色关联模型
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

// RolePermission 角色权限关联模型
type RolePermission struct {
    BaseModelWithoutUUID
    RoleID       uint64    `gorm:"not null;uniqueIndex:uk_role_permission" json:"role_id"`
    PermissionID uint64    `gorm:"not null;uniqueIndex:uk_role_permission" json:"permission_id"`
    GrantedBy    uint64    `gorm:"index" json:"granted_by"`
    GrantedAt    time.Time `gorm:"autoCreateTime" json:"granted_at"`
}
```

### 4. 字段权限模型（实际实现）

```go
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

// RoleFieldPermission 角色字段权限表
type RoleFieldPermission struct {
    BaseModelWithoutUUID
    TenantID       uint64 `gorm:"not null;index" json:"tenant_id"`
    RoleID         uint64 `gorm:"not null;index" json:"role_id"`
    EntityTable    string `gorm:"type:varchar(100);not null;index" json:"entity_table"`
    FieldName      string `gorm:"type:varchar(100);not null;index" json:"field_name"`
    PermissionType string `gorm:"type:varchar(20);not null" json:"permission_type"` // default, hidden, readonly
}

// 字段权限类型常量
const (
    FieldPermissionDefault  = "default"  // 默认：正常显示和编辑
    FieldPermissionHidden   = "hidden"   // 隐藏：不显示该字段
    FieldPermissionReadonly = "readonly" // 只读：显示但不能编辑
)
```

## 🛡️ 权限验证机制

### 1. 权限服务接口（实际实现）

```go
// PermissionService 权限服务接口
type PermissionService interface {
    // 检查用户是否拥有指定权限
    CheckUserPermission(ctx context.Context, userID, tenantID, permissionCode string) (bool, error)
    // 获取用户角色
    GetUserRoles(ctx context.Context, userID, tenantID string) ([]models.Role, error)
    // 获取用户权限
    GetUserPermissions(ctx context.Context, userID, tenantID string) ([]models.Permission, error)
    // 检查用户是否拥有指定角色
    HasRole(ctx context.Context, userID, tenantID, roleCode string) (bool, error)
    // 检查用户是否为系统管理员
    IsSystemAdmin(ctx context.Context, userID string) (bool, error)
    // 检查用户是否为租户管理员
    IsTenantAdmin(ctx context.Context, userID, tenantID string) (bool, error)
}
```

### 2. 权限验证中间件（实际实现）

```go
// RequirePermission 要求特定权限
func (m *AuthMiddleware) RequirePermission(permissionCode string) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := c.Request.Context()
        
        // 获取用户信息
        userID, exists := c.Get("user_id")
        tenantID, exists := c.Get("tenant_id")
        
        // 检查权限
        hasPermission, err := m.permissionService.CheckUserPermission(ctx, userIDStr, tenantIDStr, permissionCode)
        if err != nil || !hasPermission {
            m.responseWriter.Error(c, errors.ErrUserPermissionError())
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// RequireAnyPermission 要求任意一个权限（OR逻辑）
func (m *AuthMiddleware) RequireAnyPermission(permissionCodes ...string) gin.HandlerFunc {
    // 检查是否拥有任意一个权限
    for _, permissionCode := range permissionCodes {
        hasPermission, err := m.permissionService.CheckUserPermission(ctx, userIDStr, tenantIDStr, permissionCode)
        if err == nil && hasPermission {
            c.Next()
            return
        }
    }
    // 没有任何权限时拒绝访问
}

// RequireAllPermissions 要求所有权限（AND逻辑）
func (m *AuthMiddleware) RequireAllPermissions(permissionCodes ...string) gin.HandlerFunc {
    // 检查是否拥有所有权限
}

// RequireOwnerOrPermission 要求资源所有者或特定权限
func (m *AuthMiddleware) RequireOwnerOrPermission(resourceUserIDParam string, permissionCode string) gin.HandlerFunc {
    // 如果是资源所有者，直接允许；否则检查权限
}

// ValidateAPIPermission API权限验证中间件（支持动态路由匹配）
func (m *AuthMiddleware) ValidateAPIPermission() gin.HandlerFunc {
    // 根据请求路径和方法生成权限代码并验证
}
```

### 3. 字段权限中间件（实际实现）

```go
// InjectFieldPermissions 注入字段权限信息到上下文
func (pm *PermissionMiddleware) InjectFieldPermissions(tableName string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 获取用户字段权限
        permissions, err := pm.fieldPermissionService.GetUserFieldPermissions(ctx, userIDStr, tenantIDStr, tableName)
        
        // 将字段权限注入到上下文
        c.Set("field_permissions", permissions)
        c.Set("field_permissions_table", tableName)
        
        c.Next()
    }
}

// HasFieldPermission 检查是否有指定字段的权限
func HasFieldPermission(c *gin.Context, fieldName, requiredPermission string) bool {
    permissions, exists := GetFieldPermissions(c)
    if !exists {
        return true // 如果没有字段权限信息，默认允许
    }
    
    permission, exists := permissions[fieldName]
    if !exists {
        return true // 如果字段没有权限配置，默认允许
    }
    
    // 权限级别：default > readonly > hidden
    switch requiredPermission {
    case "default":
        return permission == "default"
    case "readonly":
        return permission == "default" || permission == "readonly"
    case "hidden":
        return true // hidden权限总是可以访问（内部使用）
    default:
        return permission == "default"
    }
}
```

### 4. 路由权限配置（实际实现）

```go
// 用户管理路由配置示例
users := api.Group("/users")
users.Use(authMiddleware.RequireAuth()) // 要求认证
{
    users.GET("", authMiddleware.RequirePermission("user_list_api"), userHandler.ListUsers)
    users.GET("/:uuid", authMiddleware.RequireOwnerOrPermission("uuid", "user_list_api"), userHandler.GetUser)
    users.PUT("/:uuid", authMiddleware.RequireOwnerOrPermission("uuid", "user_update_api"), userHandler.UpdateUser)
    users.DELETE("/:uuid", authMiddleware.RequirePermission("user_delete_api"), userHandler.DeleteUser)
}

// 管理员路由配置示例
admin := api.Group("/admin")
admin.Use(authMiddleware.RequireAuth()) // 要求认证
{
    admin.POST("/users", authMiddleware.RequirePermission("user_create_api"), userHandler.CreateUser)
}

// 角色管理路由配置示例
roles := api.Group("/roles")
roles.Use(authMiddleware.RequireAuth()) // 要求认证
{
    roles.GET("", authMiddleware.RequirePermission("role_list_api"), roleHandler.ListRoles)
    roles.POST("", authMiddleware.RequirePermission("role_create_api"), roleHandler.CreateRole)
    roles.PUT("/:id", authMiddleware.RequirePermission("role_update_api"), roleHandler.UpdateRole)
    roles.DELETE("/:id", authMiddleware.RequirePermission("role_delete_api"), roleHandler.DeleteRole)
    roles.POST("/:id/permissions", authMiddleware.RequirePermission("role_assign_api"), roleHandler.AssignPermissions)
}

// 字段权限管理路由配置示例
fieldPermissions := api.Group("/field-permissions")
fieldPermissions.Use(authMiddleware.RequireAuth()) // 要求认证
{
    fieldPermissions.GET("/tables/:tableName/fields", authMiddleware.RequirePermission("field_permission_list_api"), fieldPermissionHandler.GetTableFields)
    fieldPermissions.GET("/roles/:roleId/:tableName", authMiddleware.RequirePermission("field_permission_list_api"), fieldPermissionHandler.GetRoleFieldPermissions)
    fieldPermissions.PUT("/roles/:roleId/:tableName", authMiddleware.RequirePermission("field_permission_update_api"), fieldPermissionHandler.UpdateRoleFieldPermissions)
}
```

## 🏢 权限层级体系（实际实现）

### 1. 权限层级结构

系统采用三层权限结构，通过 `parent_code` 建立层级关系：

```
用户管理模块 (user_menu)
├── 用户列表 (user_list_btn)
│   └── 用户列表API (user_list_api) - GET /api/v1/users
├── 创建用户 (user_create_btn)
│   └── 创建用户API (user_create_api) - POST /api/v1/users
├── 编辑用户 (user_update_btn)
│   └── 更新用户API (user_update_api) - PUT /api/v1/users/:id
└── 删除用户 (user_delete_btn)
    └── 删除用户API (user_delete_api) - DELETE /api/v1/users/:id

角色管理模块 (role_menu)
├── 角色列表 (role_list_btn)
│   └── 角色列表API (role_list_api) - GET /api/v1/roles
├── 创建角色 (role_create_btn)
│   └── 创建角色API (role_create_api) - POST /api/v1/roles
├── 分配权限 (role_assign_btn)
│   └── 分配权限API (role_assign_api) - POST /api/v1/roles/:id/permissions
└── 字段权限配置 (field_permission_btn)
    ├── 字段权限列表API (field_permission_list_api) - GET /api/v1/roles/:id/field-permissions/:table
    └── 更新字段权限API (field_permission_update_api) - PUT /api/v1/roles/:id/field-permissions/:table

系统管理模块 (system_menu) - 仅系统管理员
├── 租户管理 (tenant_menu)
│   ├── 租户列表 (tenant_list_btn)
│   │   └── 租户列表API (tenant_list_api) - GET /api/v1/system/tenants
│   ├── 创建租户 (tenant_create_btn)
│   │   └── 创建租户API (tenant_create_api) - POST /api/v1/system/tenants
│   └── 更新租户 (tenant_update_btn)
│       └── 更新租户API (tenant_update_api) - PUT /api/v1/system/tenants/:id
└── 权限管理 (permission_menu)
    ├── 权限列表API (permission_list_api) - GET /api/v1/system/permissions
    └── 更新权限API (permission_update_api) - PUT /api/v1/system/permissions/:id
```

### 2. 系统角色权限配置

```go
// 系统管理员角色权限
var SystemAdminPermissions = []string{
    "system_menu", "tenant_menu", "tenant_list_btn", "tenant_list_api",
    "tenant_create_btn", "tenant_create_api", "tenant_update_btn", "tenant_update_api",
    "tenant_delete_btn", "tenant_delete_api", "permission_menu", "permission_list_api",
    "permission_update_api",
}

// 租户管理员角色权限
var TenantAdminPermissions = []string{
    "user_menu", "user_list_btn", "user_list_api", "user_create_btn", "user_create_api",
    "user_update_btn", "user_update_api", "user_delete_btn", "user_delete_api",
    "role_menu", "role_list_btn", "role_list_api", "role_create_btn", "role_create_api",
    "role_assign_btn", "role_assign_api", "field_permission_btn", "field_permission_list_api",
    "field_permission_update_api",
}
```

## 🔄 字段权限系统

### 1. 字段权限服务接口

```go
// FieldPermissionService 字段权限服务接口
type FieldPermissionService interface {
    // 获取表的字段配置
    GetTableFields(ctx context.Context, tableName string) ([]models.FieldPermission, error)
    // 获取角色的字段权限
    GetRoleFieldPermissions(ctx context.Context, roleID uint64, tableName string) ([]models.RoleFieldPermission, error)
    // 更新角色的字段权限
    UpdateRoleFieldPermissions(ctx context.Context, roleID uint64, tableName string, permissions []models.RoleFieldPermission) error
    // 获取用户的字段权限
    GetUserFieldPermissions(ctx context.Context, userID, tenantID, tableName string) (map[string]string, error)
    // 初始化表的字段权限配置
    InitializeFieldPermissions(ctx context.Context, tableName string, fields []dto.FieldConfig) error
}
```

### 2. 字段权限配置示例

```go
// 用户表字段权限配置示例
UserFieldPermissions := []models.FieldPermission{
    {
        EntityTable:  "users",
        FieldName:    "name",
        FieldLabel:   "姓名",
        FieldType:    "string",
        DefaultValue: models.FieldPermissionDefault,
    },
    {
        EntityTable:  "users",
        FieldName:    "email",
        FieldLabel:   "邮箱",
        FieldType:    "email",
        DefaultValue: models.FieldPermissionDefault,
    },
    {
        EntityTable:  "users",
        FieldName:    "phone",
        FieldLabel:   "手机号",
        FieldType:    "phone",
        DefaultValue: models.FieldPermissionDefault,
    },
    {
        EntityTable:  "users",
        FieldName:    "salary",
        FieldLabel:   "薪资",
        FieldType:    "decimal",
        DefaultValue: models.FieldPermissionHidden, // 默认隐藏敏感信息
    },
}
```

## 📊 权限初始化和迁移

### 1. 权限初始化脚本

权限和角色的初始化通过 `cmd/migrate/permissions.go` 完成：

```bash
# 运行权限初始化
go run cmd/migrate/main.go -action=permissions
```

### 2. 数据库迁移

```go
// 自动迁移所有权限相关表
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &models.Tenant{},
        &models.User{},
        &models.Permission{},
        &models.Role{},
        &models.UserRole{},
        &models.RolePermission{},
        &models.FieldPermission{},
        &models.RoleFieldPermission{},
        &models.PermissionAuditLog{},
    )
}
```

## 📈 权限审计

### 1. 权限操作审计

```go
// PermissionAuditLog 权限操作审计日志
type PermissionAuditLog struct {
    BaseModel
    TenantID     uint64 `gorm:"not null;index" json:"tenant_id"`
    OperatorID   uint64 `gorm:"not null;index" json:"operator_id"`       // 操作人
    TargetType   string `gorm:"type:varchar(50);not null" json:"target_type"` // user, role, permission
    TargetID     uint64 `gorm:"not null;index" json:"target_id"`         // 目标ID
    Action       string `gorm:"type:varchar(50);not null" json:"action"` // grant, revoke, create, delete
    Permission   string `gorm:"type:varchar(100)" json:"permission"`     // 权限代码
    OldValue     string `gorm:"type:text" json:"old_value"`              // 变更前值
    NewValue     string `gorm:"type:text" json:"new_value"`              // 变更后值
    Reason       string `gorm:"type:text" json:"reason"`                 // 操作原因
    IPAddress    string `gorm:"type:varchar(45)" json:"ip_address"`      // 操作IP
    UserAgent    string `gorm:"type:text" json:"user_agent"`             // 用户代理
}
```

## 🎯 API接口设计

### 1. 权限管理API

```http
# 获取权限列表
GET /api/v1/permissions
Authorization: Bearer <token>

# 获取权限树结构
GET /api/v1/permissions/tree
Authorization: Bearer <token>

# 系统权限管理（仅系统管理员）
GET /api/v1/system/permissions
PUT /api/v1/system/permissions/:id
Authorization: Bearer <token>
```

### 2. 角色管理API

```http
# 获取角色列表
GET /api/v1/roles
Authorization: Bearer <token>

# 创建角色
POST /api/v1/roles
Content-Type: application/json
{
    "code": "custom_role",
    "name": "自定义角色",
    "description": "角色描述"
}

# 分配权限给角色
POST /api/v1/roles/:id/permissions
Content-Type: application/json
{
    "permission_ids": ["1", "2", "3"]
}

# 获取角色权限
GET /api/v1/roles/:id/permissions
```

### 3. 字段权限管理API

```http
# 获取表字段配置
GET /api/v1/field-permissions/tables/:tableName/fields
Authorization: Bearer <token>

# 获取角色字段权限
GET /api/v1/field-permissions/roles/:roleId/:tableName
Authorization: Bearer <token>

# 更新角色字段权限
PUT /api/v1/field-permissions/roles/:roleId/:tableName
Content-Type: application/json
{
    "permissions": [
        {
            "field_name": "name",
            "permission_type": "default"
        },
        {
            "field_name": "salary",
            "permission_type": "hidden"
        }
    ]
}
```

## 🎯 最佳实践

### 1. 权限设计原则
- **最小权限原则**: 用户只拥有完成工作所需的最小权限
- **层级权限管理**: 通过菜单→按钮→API的层级结构组织权限
- **职责分离**: 敏感操作需要多个角色协作完成
- **多租户隔离**: 租户间权限完全隔离

### 2. 性能优化建议
- 使用权限缓存减少数据库查询
- 批量权限检查减少API调用
- 字段权限预加载避免N+1查询问题
- 使用动态API权限验证减少硬编码

### 3. 安全建议
- 权限变更需要审批流程
- 记录所有权限操作的审计日志
- 定期审查和清理不必要的权限
- 实施权限最小化原则
- 敏感字段默认隐藏，需要明确授权

## 🎯 总结

UltraFit权限控制系统提供了完整的企业级RBAC解决方案：

1. **完整的权限层级**: 支持菜单、按钮、API和字段四级权限控制
2. **灵活的角色管理**: 支持系统角色和自定义角色
3. **多租户隔离**: 租户级的权限和角色管理
4. **字段级权限**: 支持字段的显示/隐藏/只读控制
5. **动态权限验证**: 支持基于路径的动态API权限验证
6. **完整的审计**: 详细的权限操作记录和审计
7. **数据库驱动**: 所有权限配置存储在数据库，支持动态管理

该系统为UltraFit平台提供了安全可靠的权限控制基础，支持复杂的企业级权限管理需求。