# UltraFit 架构设计与开发规范

本文档详细描述了 UltraFit 项目的架构设计原则、分层规范和开发约束，是所有开发人员必须遵循的核心指南。

## 🏗️ 核心架构原则

### 清洁架构分层

UltraFit 严格遵循清洁架构模式，确保各层职责清晰、低耦合高内聚：

```
HTTP Request
    ↓
┌─────────────────┐
│  Handler Layer  │ ← HTTP请求处理、参数绑定、响应格式化
└─────────────────┘
    ↓ (仅调用Service)
┌─────────────────┐
│  Service Layer  │ ← 业务逻辑、事务管理、用例实现
└─────────────────┘
    ↓ (调用Repository)
┌─────────────────┐
│Repository Layer │ ← 数据访问抽象、数据库操作
└─────────────────┘
    ↓
┌─────────────────┐
│   Model Layer   │ ← 数据模型、领域实体
└─────────────────┘
    ↓
   Database
```

### 架构核心规则

#### 1. 分层依赖原则
- **Handler → Service → Repository** 单向依赖
- **禁止跨层直接调用**：Handler 不能直接调用 Repository
- **接口驱动**：所有跨层通信必须通过接口
- **依赖注入**：使用 Wire 进行自动依赖注入

#### 2. 职责分离原则
```go
// ❌ 错误：Handler 包含业务逻辑
func (h *UserHandler) CreateUser(c *gin.Context) {
    // 业务逻辑应该在 Service 中
    if user.Age < 18 {
        return errors.New("用户年龄不符合要求")
    }
    h.userRepo.Create(user) // 违规：直接调用 Repository
}

// ✅ 正确：Handler 只处理 HTTP 层
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, "参数错误", err)
        return
    }
    
    user := req.ToModel()
    if err := h.userService.CreateUser(c.Request.Context(), user); err != nil {
        response.Error(c, "创建用户失败", err)
        return
    }
    
    response.Success(c, "创建成功", user)
}
```

## 🏛️ 分层详细规范

### Handler Layer（处理器层）

**职责**：HTTP 请求处理、参数绑定、响应格式化

**规范要求**：
- 只处理 HTTP 相关逻辑，不包含业务逻辑
- 使用 Gin 框架进行参数绑定和验证
- 统一使用 `pkg/response` 进行响应格式化
- 正确传递 Context 到 Service 层
- 实现统一的错误处理

**标准实现模式**：
```go
type UserHandler struct {
    userService services.UserService
}

func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    user, err := h.userService.GetUserByID(c.Request.Context(), userID)
    if err != nil {
        response.Error(c, "获取用户失败", err)
        return
    }
    
    response.Success(c, "获取成功", user)
}
```

### Service Layer（服务层）

**职责**：业务逻辑实现、事务管理、用例编排

**规范要求**：
- 实现核心业务逻辑和业务规则验证
- 协调多个 Repository 的数据操作
- 管理数据库事务
- 处理业务异常并返回有意义的错误信息
- 所有 Service 必须定义接口

**标准实现模式**：
```go
// 接口定义
type UserService interface {
    CreateUser(ctx context.Context, user *models.User) error
    GetUserByID(ctx context.Context, id string) (*models.User, error)
}

// 实现
type userService struct {
    userRepo repositories.UserRepository
    roleRepo repositories.RoleRepository
    logger   logger.Logger
}

func (s *userService) CreateUser(ctx context.Context, user *models.User) error {
    // 业务规则验证
    if err := s.validateUserRules(user); err != nil {
        return errors.Wrap(err, "用户验证失败")
    }
    
    // 数据操作
    if err := s.userRepo.Create(ctx, user); err != nil {
        return errors.Wrap(err, "创建用户失败")
    }
    
    // 记录日志
    s.logger.InfoWithTrace(ctx, "用户创建成功", 
        zap.String("user_id", user.ID),
        zap.String("email", user.Email))
    
    return nil
}
```

### Repository Layer（仓储层）

**职责**：数据访问抽象、数据库操作封装

**规范要求**：
- 封装所有数据库操作，提供统一的数据访问接口
- 使用 GORM 进行数据库操作
- 处理数据库错误并转换为业务异常
- 实现数据查询的各种过滤、排序、分页
- 所有 Repository 必须定义接口

**标准实现模式**：
```go
// 接口定义
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id string) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
}

// 实现
type userRepository struct {
    db *gorm.DB
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
    if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
        return errors.Wrap(err, "数据库创建用户失败")
    }
    return nil
}
```

### Model Layer（模型层）

**职责**：数据模型定义、领域实体

**规范要求**：
- 定义数据库表结构和关系
- 实现模型验证规则
- 包含领域相关的方法
- 支持多租户的 `tenant_id` 字段

**标准实现模式**：
```go
type User struct {
    ID        string    `gorm:"primarykey" json:"id"`
    TenantID  string    `gorm:"not null;index" json:"tenant_id"`
    Email     string    `gorm:"uniqueIndex;not null" json:"email"`
    Name      string    `gorm:"not null" json:"name"`
    Password  string    `gorm:"not null" json:"-"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// 领域方法
func (u *User) ValidatePassword(password string) bool {
    return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}
```

## 🔌 依赖注入架构

### Wire 依赖注入规范

UltraFit 使用 Google Wire 进行依赖注入，采用模块化的 Provider 组织方式：

**Provider 文件组织**：
```
internal/
├── infrastructure/providers.go    # 基础设施：DB、Logger、Config
├── repositories/providers.go      # 数据访问层
├── services/providers.go         # 业务逻辑层
├── handlers/providers.go         # HTTP 处理层
└── middleware/providers.go       # 中间件
```

**标准 Provider 实现**：
```go
// services/providers.go
//go:build wireinject
// +build wireinject

package services

import "github.com/google/wire"

// ServicesProviderSet 服务层依赖注入
var ServicesProviderSet = wire.NewSet(
    NewUserService,
    wire.Bind(new(UserService), new(*userService)),
    
    NewAuthService,
    wire.Bind(new(AuthService), new(*authService)),
)

// 构造函数返回接口类型
func NewUserService(repo repositories.UserRepository) UserService {
    return &userService{userRepo: repo}
}
```

**关键规则**：
- 所有 Provider 函数必须返回接口类型
- 使用 `wire.Bind` 绑定接口到实现
- 修改构造函数后必须运行 `make wire`
- 避免循环依赖

## 🔒 安全架构

### 多租户安全模型

**租户隔离机制**：
- 所有用户数据包含 `tenant_id` 字段
- JWT Token 携带租户上下文信息
- 数据库查询自动过滤租户数据
- 权限系统基于租户边界运作

**权限控制层次**：
```
1. 菜单权限    - 页面访问控制
2. 按钮权限    - 操作权限控制  
3. API权限     - 接口访问控制
4. 字段权限    - 数据字段访问控制
```

### 认证与授权

**JWT 认证流程**：
```go
// 中间件验证 JWT
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        claims, err := validateJWT(token)
        if err != nil {
            response.Unauthorized(c, "认证失败")
            c.Abort()
            return
        }
        
        // 设置租户上下文
        c.Set("tenant_id", claims.TenantID)
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

**验证码安全机制**：
- 图形验证码防暴力破解
- Redis 分布式存储，内存备份
- 一次性使用，自动过期
- 登录失败计数保护

## 🧪 测试架构

### 测试策略层次

```
E2E Tests           # 端到端集成测试
    ↓
Integration Tests   # 服务间集成测试  
    ↓
Unit Tests         # 单元测试（每层独立）
```

### 测试实现规范

**单元测试模式**：
```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        user    *models.User
        mockFn  func(*mocks.UserRepository)
        wantErr bool
    }{
        {
            name: "成功创建用户",
            user: &models.User{Email: "test@example.com"},
            mockFn: func(repo *mocks.UserRepository) {
                repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
            },
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试实现
        })
    }
}
```

**覆盖率要求**：
- 单元测试覆盖率 > 80%
- 核心业务逻辑必须 100% 覆盖
- 接口层必须有集成测试

## 📊 可观测性架构

### OpenTelemetry 集成

**追踪链路**：
```
HTTP Request → Handler → Service → Repository → Database
      ↓           ↓         ↓           ↓
   Trace ID   Span ID   Span ID    Span ID
```

**实现模式**：
```go
func (s *userService) CreateUser(ctx context.Context, user *models.User) error {
    // 创建 Span
    ctx, span := trace.StartSpan(ctx, "userService.CreateUser")
    defer span.End()
    
    // 记录关键属性
    span.SetAttributes(
        attribute.String("user.email", user.Email),
        attribute.String("tenant.id", user.TenantID),
    )
    
    // 业务逻辑实现
    if err := s.userRepo.Create(ctx, user); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }
    
    return nil
}
```

### 结构化日志

**日志规范**：
```go
// 使用 TraceID 关联日志
logger.InfoWithTrace(ctx, "用户创建成功",
    zap.String("user_id", user.ID),
    zap.String("tenant_id", user.TenantID),
    zap.String("operation", "create_user"),
)
```

## 🔧 开发工作流

### 新功能开发流程

1. **模型定义** (`internal/models/`)
   ```go
   type NewEntity struct {
       ID       string `gorm:"primarykey"`
       TenantID string `gorm:"not null;index"`
       // 其他字段...
   }
   ```

2. **Repository 实现** (`internal/repositories/`)
   ```go
   type NewEntityRepository interface {
       Create(ctx context.Context, entity *models.NewEntity) error
   }
   ```

3. **Service 实现** (`internal/services/`)
   ```go
   type NewEntityService interface {
       CreateEntity(ctx context.Context, entity *models.NewEntity) error
   }
   ```

4. **Handler 实现** (`internal/handlers/`)
   ```go
   func (h *NewEntityHandler) Create(c *gin.Context) {
       // HTTP 处理逻辑
   }
   ```

5. **路由注册** (`internal/routes/`)
   ```go
   router.POST("/entities", handlers.NewEntity.Create)
   ```

6. **依赖注入** (`*/providers.go`)
   ```go
   // 添加到相应的 ProviderSet
   ```

7. **Wire 生成**
   ```bash
   make wire  # 重新生成依赖注入代码
   ```

### 代码审查清单

**Handler 层检查**：
- [ ] 是否只处理 HTTP 层逻辑？
- [ ] 是否正确使用参数绑定？
- [ ] 是否实现统一错误处理？
- [ ] 是否正确传递 Context？

**Service 层检查**：
- [ ] 是否实现接口定义？
- [ ] 是否包含业务逻辑验证？
- [ ] 是否正确处理事务？
- [ ] 是否添加操作日志？

**Repository 层检查**：
- [ ] 是否使用 Repository 模式？
- [ ] 是否正确处理 GORM 错误？
- [ ] 是否避免 N+1 查询？
- [ ] 是否使用参数化查询？

**安全检查**：
- [ ] 是否遵循分层架构？
- [ ] 是否实现租户隔离？
- [ ] 是否正确验证用户输入？
- [ ] 是否包含权限验证？

## 📚 相关文档

- 📖 [快速开始指南](./getting-started.md) - 环境搭建和项目初始化
- 🔧 [API 开发指南](./api-guide.md) - API 设计和实现规范
- 🧪 [测试指南](./testing-guide.md) - 测试策略和实现

## 💡 架构决策记录

### ADR-001: 选择清洁架构
**决策**：采用清洁架构模式进行分层设计  
**理由**：确保高可测试性、低耦合、易维护  
**影响**：严格的分层约束，但提高了代码质量

### ADR-002: 使用 Wire 依赖注入
**决策**：使用 Google Wire 进行依赖注入  
**理由**：编译时生成，无运行时开销，类型安全  
**影响**：需要学习 Wire 语法，但提供了更好的性能

### ADR-003: 多租户数据隔离
**决策**：在数据层实现租户隔离  
**理由**：确保数据安全，支持 SaaS 模式  
**影响**：所有数据模型需要包含 tenant_id

---

**重要提醒**：这些架构规范是项目代码质量的基石，所有开发人员都必须严格遵守。违反架构约束的代码将不会通过代码审查。