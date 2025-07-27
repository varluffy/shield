# 数据库模型设计

## 🎯 设计目标

UltraFit数据库设计基于多租户架构，采用行级隔离模式，支持认证、权限控制和业务功能。

## 🏗️ 数据库架构

### 1. 多租户数据隔离

- 所有业务表都包含 `tenant_id` 字段
- 通过应用层确保数据隔离
- 索引策略：`(tenant_id, other_columns)`

### 2. 核心设计原则

- **数据隔离**: 每个业务表都有 tenant_id 字段
- **软删除**: 使用 deleted_at 字段实现软删除
- **审计字段**: created_at, updated_at 自动维护
- **唯一约束**: 租户内的唯一性约束

## 📊 核心数据表设计

### 1. 租户表 (tenants)

```sql
CREATE TABLE tenants (
    id VARCHAR(36) PRIMARY KEY COMMENT '租户ID',
    name VARCHAR(100) NOT NULL COMMENT '租户名称',
    domain VARCHAR(100) UNIQUE COMMENT '域名',
    status ENUM('active', 'inactive', 'suspended') DEFAULT 'active' COMMENT '状态',
    plan VARCHAR(50) DEFAULT 'basic' COMMENT '套餐类型',
    max_users INT DEFAULT 100 COMMENT '最大用户数',
    max_storage BIGINT DEFAULT 1073741824 COMMENT '最大存储空间(字节)',
    settings JSON COMMENT '租户配置',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    
    INDEX idx_domain (domain),
    INDEX idx_status (status),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='租户表';
```

### 2. 用户表 (users)

```sql
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY COMMENT '用户ID',
    tenant_id VARCHAR(36) NOT NULL COMMENT '租户ID',
    email VARCHAR(255) NOT NULL COMMENT '邮箱',
    password VARCHAR(255) NOT NULL COMMENT '密码哈希',
    name VARCHAR(100) COMMENT '姓名',
    avatar VARCHAR(500) COMMENT '头像URL',
    phone VARCHAR(20) COMMENT '手机号',
    status ENUM('active', 'inactive', 'locked') DEFAULT 'active' COMMENT '状态',
    email_verified_at TIMESTAMP NULL COMMENT '邮箱验证时间',
    last_login_at TIMESTAMP NULL COMMENT '最后登录时间',
    login_count INT DEFAULT 0 COMMENT '登录次数',
    failed_login_attempts INT DEFAULT 0 COMMENT '失败登录次数',
    locked_until TIMESTAMP NULL COMMENT '锁定到期时间',
    timezone VARCHAR(50) DEFAULT 'Asia/Shanghai' COMMENT '时区',
    language VARCHAR(10) DEFAULT 'zh' COMMENT '语言',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    
    -- 租户内邮箱唯一
    UNIQUE KEY uk_tenant_email (tenant_id, email),
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_email (email),
    INDEX idx_status (status),
    INDEX idx_deleted_at (deleted_at),
    
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='用户表';
```

### 3. 权限表 (permissions)

```sql
CREATE TABLE permissions (
    id VARCHAR(36) PRIMARY KEY COMMENT '权限ID',
    code VARCHAR(100) NOT NULL UNIQUE COMMENT '权限代码',
    name VARCHAR(100) NOT NULL COMMENT '权限名称',
    description TEXT COMMENT '权限描述',
    resource VARCHAR(100) NOT NULL COMMENT '资源类型',
    action VARCHAR(50) NOT NULL COMMENT '操作类型',
    type ENUM('function', 'data') DEFAULT 'function' COMMENT '权限类型',
    category VARCHAR(50) COMMENT '权限分类',
    is_system BOOLEAN DEFAULT FALSE COMMENT '是否系统权限',
    sort_order INT DEFAULT 0 COMMENT '排序',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_code (code),
    INDEX idx_resource (resource),
    INDEX idx_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='权限表';
```

### 4. 角色表 (roles)

```sql
CREATE TABLE roles (
    id VARCHAR(36) PRIMARY KEY COMMENT '角色ID',
    tenant_id VARCHAR(36) NOT NULL COMMENT '租户ID',
    code VARCHAR(100) NOT NULL COMMENT '角色代码',
    name VARCHAR(100) NOT NULL COMMENT '角色名称',
    description TEXT COMMENT '角色描述',
    type ENUM('system', 'custom') DEFAULT 'custom' COMMENT '角色类型',
    level INT DEFAULT 0 COMMENT '角色级别',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    
    -- 租户内角色代码唯一
    UNIQUE KEY uk_tenant_code (tenant_id, code),
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_type (type),
    INDEX idx_deleted_at (deleted_at),
    
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='角色表';
```

### 5. 用户角色关联表 (user_roles)

```sql
CREATE TABLE user_roles (
    id VARCHAR(36) PRIMARY KEY COMMENT '关联ID',
    user_id VARCHAR(36) NOT NULL COMMENT '用户ID',
    role_id VARCHAR(36) NOT NULL COMMENT '角色ID',
    tenant_id VARCHAR(36) NOT NULL COMMENT '租户ID',
    granted_by VARCHAR(36) COMMENT '授权人ID',
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',
    expires_at TIMESTAMP NULL COMMENT '过期时间',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    
    -- 用户角色唯一
    UNIQUE KEY uk_user_role (user_id, role_id),
    INDEX idx_user_id (user_id),
    INDEX idx_role_id (role_id),
    INDEX idx_tenant_id (tenant_id),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='用户角色关联表';
```

### 6. 角色权限关联表 (role_permissions)

```sql
CREATE TABLE role_permissions (
    id VARCHAR(36) PRIMARY KEY COMMENT '关联ID',
    role_id VARCHAR(36) NOT NULL COMMENT '角色ID',
    permission_id VARCHAR(36) NOT NULL COMMENT '权限ID',
    granted_by VARCHAR(36) COMMENT '授权人ID',
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',
    
    -- 角色权限唯一
    UNIQUE KEY uk_role_permission (role_id, permission_id),
    INDEX idx_role_id (role_id),
    INDEX idx_permission_id (permission_id),
    
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='角色权限关联表';
```

## 🔐 认证相关表

### 1. 刷新令牌表 (refresh_tokens)

```sql
CREATE TABLE refresh_tokens (
    id VARCHAR(36) PRIMARY KEY COMMENT '令牌ID',
    user_id VARCHAR(36) NOT NULL COMMENT '用户ID',
    tenant_id VARCHAR(36) NOT NULL COMMENT '租户ID',
    token_hash VARCHAR(255) NOT NULL UNIQUE COMMENT '令牌哈希',
    expires_at TIMESTAMP NOT NULL COMMENT '过期时间',
    is_revoked BOOLEAN DEFAULT FALSE COMMENT '是否撤销',
    user_agent TEXT COMMENT '用户代理',
    ip_address VARCHAR(45) COMMENT 'IP地址',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    INDEX idx_user_id (user_id),
    INDEX idx_token_hash (token_hash),
    INDEX idx_expires_at (expires_at),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='刷新令牌表';
```

### 2. 登录尝试记录表 (login_attempts)

```sql
CREATE TABLE login_attempts (
    id VARCHAR(36) PRIMARY KEY COMMENT '记录ID',
    user_id VARCHAR(36) COMMENT '用户ID',
    tenant_id VARCHAR(36) COMMENT '租户ID',
    email VARCHAR(255) NOT NULL COMMENT '登录邮箱',
    ip_address VARCHAR(45) NOT NULL COMMENT 'IP地址',
    user_agent TEXT COMMENT '用户代理',
    success BOOLEAN DEFAULT FALSE COMMENT '是否成功',
    failure_reason VARCHAR(100) COMMENT '失败原因',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    INDEX idx_user_id (user_id),
    INDEX idx_email (email),
    INDEX idx_ip_address (ip_address),
    INDEX idx_created_at (created_at),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='登录尝试记录表';
```

### 3. 认证事件表 (auth_events)

```sql
CREATE TABLE auth_events (
    id VARCHAR(36) PRIMARY KEY COMMENT '事件ID',
    user_id VARCHAR(36) COMMENT '用户ID',
    tenant_id VARCHAR(36) COMMENT '租户ID',
    event VARCHAR(50) NOT NULL COMMENT '事件类型',
    ip_address VARCHAR(45) COMMENT 'IP地址',
    user_agent TEXT COMMENT '用户代理',
    success BOOLEAN DEFAULT TRUE COMMENT '是否成功',
    message TEXT COMMENT '事件消息',
    metadata JSON COMMENT '元数据',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    INDEX idx_user_id (user_id),
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_event (event),
    INDEX idx_created_at (created_at),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='认证事件表';
```

## 📈 审计表

### 1. 权限审计表 (permission_audits)

```sql
CREATE TABLE permission_audits (
    id VARCHAR(36) PRIMARY KEY COMMENT '审计ID',
    tenant_id VARCHAR(36) NOT NULL COMMENT '租户ID',
    operator_id VARCHAR(36) NOT NULL COMMENT '操作人ID',
    target_id VARCHAR(36) COMMENT '目标ID',
    operation VARCHAR(50) NOT NULL COMMENT '操作类型',
    resource VARCHAR(100) NOT NULL COMMENT '资源类型',
    details JSON COMMENT '详细信息',
    ip_address VARCHAR(45) COMMENT 'IP地址',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_operator_id (operator_id),
    INDEX idx_operation (operation),
    INDEX idx_created_at (created_at),
    
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (operator_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='权限审计表';
```

## 🔧 GORM模型定义

### 1. 基础模型

```go
// 租户基础模型
type TenantModel struct {
    ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
    TenantID  string         `gorm:"type:varchar(36);not null;index" json:"tenant_id"`
    CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// 系统基础模型（无租户隔离）
type BaseModel struct {
    ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
    CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
```

### 2. 租户模型

```go
type Tenant struct {
    BaseModel
    Name       string `gorm:"type:varchar(100);not null" json:"name"`
    Domain     string `gorm:"type:varchar(100);uniqueIndex" json:"domain"`
    Status     string `gorm:"type:enum('active','inactive','suspended');default:'active'" json:"status"`
    Plan       string `gorm:"type:varchar(50);default:'basic'" json:"plan"`
    MaxUsers   int    `gorm:"default:100" json:"max_users"`
    MaxStorage int64  `gorm:"default:1073741824" json:"max_storage"`
    Settings   string `gorm:"type:json" json:"settings"`
}

func (Tenant) TableName() string {
    return "tenants"
}
```

### 3. 用户模型

```go
type User struct {
    TenantModel
    Email                string     `gorm:"type:varchar(255);not null;uniqueIndex:uk_tenant_email" json:"email"`
    Password             string     `gorm:"type:varchar(255);not null" json:"-"`
    Name                 string     `gorm:"type:varchar(100)" json:"name"`
    Avatar               string     `gorm:"type:varchar(500)" json:"avatar"`
    Phone                string     `gorm:"type:varchar(20)" json:"phone"`
    Status               string     `gorm:"type:enum('active','inactive','locked');default:'active'" json:"status"`
    EmailVerifiedAt      *time.Time `json:"email_verified_at"`
    LastLoginAt          *time.Time `json:"last_login_at"`
    LoginCount           int        `gorm:"default:0" json:"login_count"`
    FailedLoginAttempts  int        `gorm:"default:0" json:"failed_login_attempts"`
    LockedUntil          *time.Time `json:"locked_until"`
    Timezone             string     `gorm:"type:varchar(50);default:'Asia/Shanghai'" json:"timezone"`
    Language             string     `gorm:"type:varchar(10);default:'zh'" json:"language"`
    
    // 关联关系
    Roles []Role `gorm:"many2many:user_roles" json:"roles,omitempty"`
}

func (User) TableName() string {
    return "users"
}

// GORM钩子：创建前生成ID
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.ID == "" {
        u.ID = uuid.New().String()
    }
    if u.TenantID == "" {
        if tenantID := tx.Statement.Context.Value("tenant_id"); tenantID != nil {
            u.TenantID = tenantID.(string)
        }
    }
    return nil
}
```

### 4. 权限和角色模型

```go
type Permission struct {
    BaseModel
    Code        string `gorm:"type:varchar(100);not null;uniqueIndex" json:"code"`
    Name        string `gorm:"type:varchar(100);not null" json:"name"`
    Description string `gorm:"type:text" json:"description"`
    Resource    string `gorm:"type:varchar(100);not null" json:"resource"`
    Action      string `gorm:"type:varchar(50);not null" json:"action"`
    Type        string `gorm:"type:enum('function','data');default:'function'" json:"type"`
    Category    string `gorm:"type:varchar(50)" json:"category"`
    IsSystem    bool   `gorm:"default:false" json:"is_system"`
    SortOrder   int    `gorm:"default:0" json:"sort_order"`
}

func (Permission) TableName() string {
    return "permissions"
}

type Role struct {
    TenantModel
    Code        string `gorm:"type:varchar(100);not null;uniqueIndex:uk_tenant_code" json:"code"`
    Name        string `gorm:"type:varchar(100);not null" json:"name"`
    Description string `gorm:"type:text" json:"description"`
    Type        string `gorm:"type:enum('system','custom');default:'custom'" json:"type"`
    Level       int    `gorm:"default:0" json:"level"`
    IsActive    bool   `gorm:"default:true" json:"is_active"`
    
    // 关联关系
    Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
    Users       []User       `gorm:"many2many:user_roles" json:"users,omitempty"`
}

func (Role) TableName() string {
    return "roles"
}
```

### 5. 关联模型

```go
type UserRole struct {
    ID        string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
    UserID    string     `gorm:"type:varchar(36);not null;uniqueIndex:uk_user_role" json:"user_id"`
    RoleID    string     `gorm:"type:varchar(36);not null;uniqueIndex:uk_user_role" json:"role_id"`
    TenantID  string     `gorm:"type:varchar(36);not null;index" json:"tenant_id"`
    GrantedBy string     `gorm:"type:varchar(36)" json:"granted_by"`
    GrantedAt time.Time  `gorm:"autoCreateTime" json:"granted_at"`
    ExpiresAt *time.Time `json:"expires_at"`
    IsActive  bool       `gorm:"default:true" json:"is_active"`
}

func (UserRole) TableName() string {
    return "user_roles"
}

type RolePermission struct {
    ID           string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
    RoleID       string    `gorm:"type:varchar(36);not null;uniqueIndex:uk_role_permission" json:"role_id"`
    PermissionID string    `gorm:"type:varchar(36);not null;uniqueIndex:uk_role_permission" json:"permission_id"`
    GrantedBy    string    `gorm:"type:varchar(36)" json:"granted_by"`
    GrantedAt    time.Time `gorm:"autoCreateTime" json:"granted_at"`
}

func (RolePermission) TableName() string {
    return "role_permissions"
}
```

### 6. 认证相关模型

```go
type RefreshToken struct {
    ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
    UserID    string    `gorm:"type:varchar(36);not null;index" json:"user_id"`
    TenantID  string    `gorm:"type:varchar(36);not null;index" json:"tenant_id"`
    TokenHash string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"token_hash"`
    ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
    IsRevoked bool      `gorm:"default:false" json:"is_revoked"`
    UserAgent string    `gorm:"type:text" json:"user_agent"`
    IPAddress string    `gorm:"type:varchar(45)" json:"ip_address"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (RefreshToken) TableName() string {
    return "refresh_tokens"
}

type LoginAttempt struct {
    ID            string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
    UserID        string    `gorm:"type:varchar(36);index" json:"user_id"`
    TenantID      string    `gorm:"type:varchar(36);index" json:"tenant_id"`
    Email         string    `gorm:"type:varchar(255);index" json:"email"`
    IPAddress     string    `gorm:"type:varchar(45)" json:"ip_address"`
    UserAgent     string    `gorm:"type:text" json:"user_agent"`
    Success       bool      `gorm:"default:false" json:"success"`
    FailureReason string    `gorm:"type:varchar(100)" json:"failure_reason"`
    CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (LoginAttempt) TableName() string {
    return "login_attempts"
}
```

## 📊 索引优化策略

### 1. 核心索引

```sql
-- 用户表索引
CREATE INDEX idx_users_tenant_status ON users(tenant_id, status);
CREATE INDEX idx_users_tenant_email ON users(tenant_id, email);

-- 角色表索引
CREATE INDEX idx_roles_tenant_type ON roles(tenant_id, type);
CREATE INDEX idx_roles_tenant_active ON roles(tenant_id, is_active);

-- 用户角色关联表索引
CREATE INDEX idx_user_roles_tenant_active ON user_roles(tenant_id, is_active);
CREATE INDEX idx_user_roles_expires ON user_roles(expires_at);

-- 令牌表索引
CREATE INDEX idx_refresh_tokens_user_active ON refresh_tokens(user_id, is_revoked);
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at);
```

### 2. 查询优化索引

```sql
-- 权限查询优化
CREATE INDEX idx_user_permissions_query ON user_roles(user_id, tenant_id, is_active, expires_at);
CREATE INDEX idx_role_permissions_query ON role_permissions(role_id);

-- 登录相关查询优化
CREATE INDEX idx_login_attempts_email_time ON login_attempts(email, created_at);
CREATE INDEX idx_login_attempts_ip_time ON login_attempts(ip_address, created_at);
```

## 🔄 数据迁移

### 1. 初始化数据

```sql
-- 插入系统权限
INSERT INTO permissions (id, code, name, resource, action, type, category, is_system) VALUES
('perm-user-create', 'user:create', '创建用户', 'user', 'create', 'function', 'user', true),
('perm-user-read', 'user:read', '查看用户', 'user', 'read', 'function', 'user', true),
('perm-user-update', 'user:update', '更新用户', 'user', 'update', 'function', 'user', true),
('perm-user-delete', 'user:delete', '删除用户', 'user', 'delete', 'function', 'user', true),
('perm-role-manage', 'role:manage', '管理角色', 'role', 'manage', 'function', 'role', true);

-- 创建默认租户
INSERT INTO tenants (id, name, domain, status) VALUES
('default-tenant', '默认租户', 'default.ultrafit.com', 'active');

-- 创建系统角色
INSERT INTO roles (id, tenant_id, code, name, type, level) VALUES
('role-system-admin', 'default-tenant', 'system_admin', '系统管理员', 'system', 100),
('role-tenant-admin', 'default-tenant', 'tenant_admin', '租户管理员', 'system', 90),
('role-user', 'default-tenant', 'user', '普通用户', 'system', 10);
```

### 2. GORM自动迁移

```go
// 自动迁移所有表
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &Tenant{},
        &User{},
        &Permission{},
        &Role{},
        &UserRole{},
        &RolePermission{},
        &RefreshToken{},
        &LoginAttempt{},
        &AuthEvent{},
        &PermissionAudit{},
    )
}

// 初始化数据
func SeedData(db *gorm.DB) error {
    // 创建系统权限
    systemPermissions := getSystemPermissions()
    for _, perm := range systemPermissions {
        var existing Permission
        if err := db.Where("code = ?", perm.Code).First(&existing).Error; err == gorm.ErrRecordNotFound {
            if err := db.Create(&perm).Error; err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

## 🎯 数据库连接配置

```go
func SetupDatabase(config *DatabaseConfig) (*gorm.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        config.User, config.Password, config.Host, config.Port, config.Name)
    
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        NamingStrategy: schema.NamingStrategy{
            SingularTable: true, // 使用单数表名
        },
    })
    if err != nil {
        return nil, err
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    
    // 连接池配置
    sqlDB.SetMaxOpenConns(config.MaxOpenConns)     // 最大连接数
    sqlDB.SetMaxIdleConns(config.MaxIdleConns)     // 最大空闲连接数
    sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime) // 连接最大生存时间
    
    return db, nil
}
```

## 🎯 最佳实践

### 1. 数据库设计原则
- 遵循第三范式，避免数据冗余
- 合理使用外键约束保证数据一致性
- 所有业务表包含租户ID实现隔离
- 使用软删除保护重要数据

### 2. 性能优化
- 合理设计索引，避免全表扫描
- 使用复合索引优化多条件查询
- 定期清理历史数据和日志
- 监控慢查询并优化

### 3. 安全考虑
- 敏感数据加密存储
- 使用参数化查询防止SQL注入
- 定期备份数据库
- 限制数据库访问权限

## 🎯 总结

UltraFit数据库设计提供了完整的多租户支持：

1. **多租户隔离**: 行级隔离确保数据安全
2. **完整的RBAC**: 用户、角色、权限的完整模型
3. **认证支持**: JWT令牌管理和登录安全
4. **审计追踪**: 完整的操作记录和日志
5. **性能优化**: 合理的索引和连接池配置

该数据库设计为UltraFit系统提供了稳定、安全、高性能的数据存储基础。 