# 多租户架构设计

## 🏗️ 架构概述

Shield采用**行级隔离**的多租户架构模式，通过租户ID在应用层实现数据隔离，既保证了数据安全，又提供了良好的资源利用率和成本效益。

## 🎯 设计目标

### 1. 数据隔离
- **完全隔离**: 租户间数据完全隔离，无法跨租户访问
- **透明隔离**: 对业务逻辑透明，开发人员无需关心隔离细节
- **性能隔离**: 单个租户的操作不影响其他租户性能

### 2. 成本效益
- **资源共享**: 多租户共享基础设施和应用服务
- **弹性扩展**: 根据租户数量和负载动态扩展
- **运维简化**: 统一的运维管理，降低运维成本

### 3. 安全合规
- **访问控制**: 严格的租户级访问控制
- **数据加密**: 敏感数据加密存储
- **审计追踪**: 完整的操作审计日志

## 🏗️ 架构模式

### 选择：行级隔离模式

**行级隔离 (Row-Level Isolation)**
- **实现方式**: 在每张表中添加`tenant_id`字段
- **优点**: 资源利用率高、扩展性好、运维成本低
- **缺点**: 需要应用层保证隔离、查询性能可能受影响

### 备选方案对比

| 方案 | 隔离级别 | 资源利用率 | 扩展性 | 运维复杂度 | 适用场景 |
|------|----------|------------|--------|------------|----------|
| 独立数据库 | 高 | 低 | 差 | 高 | 大型企业客户 |
| 数据库分片 | 中 | 中 | 中 | 中 | 中型企业客户 |
| 行级隔离 | 中 | 高 | 好 | 低 | 中小型企业客户 |

## 🎨 技术实现

### 1. 租户上下文管理

```go
// 租户上下文
type TenantContext struct {
    TenantID   string
    TenantName string
    Domain     string
    UserID     string
    UserEmail  string
}

// 中间件：从请求中提取租户信息
func TenantMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := extractTenantFromJWT(c)
        
        if tenantID == "" {
            tenantID = extractTenantFromDomain(c)
        }
        
        if tenantID == "" {
            tenantID = c.GetHeader("X-Tenant-ID")
        }
        
        if tenantID == "" {
            c.JSON(400, gin.H{"error": "租户信息缺失"})
            c.Abort()
            return
        }
        
        ctx := context.WithValue(c.Request.Context(), "tenant_id", tenantID)
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}
```

### 2. 数据访问层隔离

```go
// 基础Repository实现
type BaseRepository struct {
    db *gorm.DB
}

func (r *BaseRepository) WithTenant(ctx context.Context) *gorm.DB {
    tenantID := ctx.Value("tenant_id").(string)
    return r.db.Where("tenant_id = ?", tenantID)
}

// 用户Repository示例
type UserRepository struct {
    BaseRepository
}

func (r *UserRepository) FindAll(ctx context.Context) ([]models.User, error) {
    var users []models.User
    err := r.WithTenant(ctx).Find(&users).Error
    return users, err
}
```

### 3. 模型设计

```go
// 租户基础模型
type TenantModel struct {
    ID        string    `gorm:"primary_key;type:varchar(36)" json:"id"`
    TenantID  string    `gorm:"index;type:varchar(36);not null" json:"tenant_id"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// 用户模型
type User struct {
    TenantModel
    Email     string `gorm:"type:varchar(255);not null;uniqueIndex:idx_tenant_email" json:"email"`
    Password  string `gorm:"type:varchar(255);not null" json:"-"`
    Name      string `gorm:"type:varchar(100)" json:"name"`
    Status    string `gorm:"type:enum('active','inactive','locked');default:'active'" json:"status"`
    Roles     []Role `gorm:"many2many:user_roles" json:"roles"`
}
```

## 🔒 安全措施

### 1. 访问控制

```go
// 租户访问控制中间件
func TenantAccessControl() gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := c.Request.Context().Value("tenant_id").(string)
        userTenantID := getUserTenantFromToken(c)
        
        if tenantID != userTenantID {
            c.JSON(403, gin.H{"error": "无权访问该租户资源"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 2. 数据查询安全

```go
// GORM Hook：自动添加租户过滤
func (db *TenantDB) BeforeCreate(tx *gorm.DB) error {
    if tenantID := tx.Statement.Context.Value("tenant_id"); tenantID != nil {
        tx.Statement.SetColumn("tenant_id", tenantID)
    }
    return nil
}

func (db *TenantDB) BeforeUpdate(tx *gorm.DB) error {
    if tenantID := tx.Statement.Context.Value("tenant_id"); tenantID != nil {
        tx.Statement.Where("tenant_id = ?", tenantID)
    }
    return nil
}
```

## 🔍 租户识别策略

### 1. 子域名识别
```
tenant1.shield.com -> tenant_id: tenant1
tenant2.shield.com -> tenant_id: tenant2
```

### 2. JWT令牌识别
```go
// JWT载荷中包含租户信息
type JWTClaims struct {
    UserID   string `json:"user_id"`
    TenantID string `json:"tenant_id"`
    Email    string `json:"email"`
    Roles    []string `json:"roles"`
    jwt.StandardClaims
}
```

### 3. 请求头识别
```http
GET /api/v1/users
X-Tenant-ID: tenant1
Authorization: Bearer <jwt_token>
```

## 🏢 系统租户设计

### 1. 系统租户概念

**系统租户（System Tenant）**是一个特殊的虚拟租户，用于管理系统级别的资源和权限：

- **租户ID**: `tenant_id = 0`
- **用途**: 存储系统管理员、系统角色、系统权限等
- **特点**: 不是真实的业务租户，纯粹用于系统管理

### 2. 系统租户与普通租户的区别

| 特性 | 系统租户 (tenant_id=0) | 普通租户 (tenant_id>0) |
|------|----------------------|----------------------|
| 数据库记录 | 无对应tenant记录 | 有对应tenant记录 |
| 用户类型 | 系统管理员 | 租户用户 |
| 权限范围 | 全系统权限 | 租户内权限 |
| 管理职责 | 管理所有租户 | 管理自己租户 |

### 3. 系统租户的应用场景

```go
// 1. 系统管理员检查
func (s *permissionService) IsSystemAdmin(ctx context.Context, userID string) (bool, error) {
    user, err := s.userRepo.GetByUUID(ctx, userID)
    if err != nil {
        return false, err
    }
    
    // 系统管理员必须属于系统租户
    if user.TenantID != 0 {
        return false, nil
    }
    
    // 检查系统管理员角色
    return s.HasRole(ctx, userID, "0", "system_admin")
}

// 2. 系统级资源初始化
func InitSystemRoles(db *gorm.DB) error {
    // 系统角色使用 tenant_id=0
    systemRole := models.Role{
        TenantModel: models.TenantModel{TenantID: 0},
        Code:        "system_admin",
        Name:        "系统管理员",
        Type:        "system",
    }
    return db.Create(&systemRole).Error
}
```

### 4. 系统租户的优势

1. **清晰的权限层级**: 系统级权限与租户级权限完全分离
2. **安全的隔离**: 系统管理员与普通租户用户完全隔离
3. **简化的实现**: 通过统一的tenant_id字段处理所有多租户逻辑
4. **灵活的扩展**: 支持未来添加更多系统级功能

## 📊 性能优化

### 1. 数据库索引
```sql
-- 租户相关索引
CREATE INDEX idx_tenant_id ON users(tenant_id);
CREATE INDEX idx_tenant_created ON users(tenant_id, created_at);
CREATE UNIQUE INDEX idx_tenant_email ON users(tenant_id, email);
```

### 2. 缓存策略
```go
// 租户级缓存
type TenantCache struct {
    redis *redis.Client
}

func (c *TenantCache) Get(tenantID, key string) (string, error) {
    cacheKey := fmt.Sprintf("tenant:%s:%s", tenantID, key)
    return c.redis.Get(context.Background(), cacheKey).Result()
}
```

## 🎯 租户管理

### 1. 租户模型
```go
type Tenant struct {
    ID          string    `gorm:"primary_key;type:varchar(36)" json:"id"`
    Name        string    `gorm:"type:varchar(100);not null" json:"name"`
    Domain      string    `gorm:"type:varchar(100);uniqueIndex" json:"domain"`
    Status      string    `gorm:"type:enum('active','inactive','suspended');default:'active'" json:"status"`
    Plan        string    `gorm:"type:varchar(50);default:'basic'" json:"plan"`
    MaxUsers    int       `gorm:"default:100" json:"max_users"`
    MaxStorage  int64     `gorm:"default:1073741824" json:"max_storage"` // 1GB
    CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
```

### 2. 配额管理
```go
// 配额检查中间件
func TenantQuotaCheck() gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := c.Request.Context().Value("tenant_id").(string)
        
        // 检查用户数量配额
        if c.Request.Method == "POST" && strings.Contains(c.Request.URL.Path, "/users") {
            if err := checkUserQuota(tenantID); err != nil {
                c.JSON(429, gin.H{"error": "用户数量超过配额"})
                c.Abort()
                return
            }
        }
        
        c.Next()
    }
}
```

## 📋 最佳实践

### 1. 开发规范
- 所有数据访问都必须通过Repository层
- 每个查询都必须包含租户ID过滤
- 使用Context传递租户信息
- 定期审查数据访问代码

### 2. 测试策略
- 单元测试必须包含租户隔离测试
- 集成测试验证跨租户数据隔离
- 性能测试评估多租户负载
- 安全测试验证租户访问控制

### 3. 监控告警
- 监控租户资源使用情况
- 监控跨租户访问尝试
- 监控数据库查询性能
- 监控缓存命中率

## 🎯 总结

多租户架构是UltraFit系统的核心基础，通过行级隔离模式实现了数据安全、成本效益和运维简化的平衡。关键实现要点：

1. **租户上下文管理**: 通过中间件自动提取和传递租户信息
2. **数据访问隔离**: 在Repository层自动添加租户过滤
3. **安全访问控制**: 多层次的租户访问验证
4. **性能优化**: 合理的索引和缓存策略
5. **监控运维**: 完善的监控和管理工具

这个架构为后续的认证系统和权限控制奠定了坚实的基础。 