# Go Web框架最佳实践指南

UltraFit项目采用Gin + Wire的组合，本指南整合了Web开发和依赖注入的最佳实践。

## 🚀 Gin Web框架核心原则

### 路由设计
- 使用**路由组(Router Groups)**组织相关的端点，提高代码可维护性
- 按功能模块或API版本组织路由：`v1.GET("/users", handler.GetUsers)`
- 使用RESTful路由约定，保持URL设计的一致性
- 避免深层嵌套路由，保持路径简洁明了

### Handler层设计
```go
// ✅ 好的做法：Handler只处理HTTP层逻辑
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }
    
    user, err := h.userService.CreateUser(c.Request.Context(), req)
    if err != nil {
        handleError(c, err)
        return
    }
    
    c.JSON(http.StatusCreated, SuccessResponse{Data: user})
}

// ❌ 避免：在Handler中处理业务逻辑
func (h *UserHandler) CreateUser(c *gin.Context) {
    // 不要在这里写复杂的业务逻辑
    // 不要在这里直接操作数据库
}
```

## 🛡️ 中间件最佳实践

### 中间件链设计
```go
// 推荐的中间件顺序
r := gin.Default()
r.Use(
    middleware.CORS(),           // CORS处理
    middleware.RequestID(),      // 请求ID生成
    middleware.Logger(),         // 请求日志
    middleware.Recovery(),       // 恐慌恢复
    middleware.RateLimit(),      // 限流
    middleware.Auth(),           // 认证
    middleware.Permission(),     // 权限检查
)
```

### 自定义中间件
```go
// ✅ 好的中间件设计
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
            param.ClientIP,
            param.TimeStamp.Format(time.RFC1123),
            param.Method,
            param.Path,
            param.Request.Proto,
            param.StatusCode,
            param.Latency,
            param.Request.UserAgent(),
            param.ErrorMessage,
        )
    })
}
```

## 🔧 Wire依赖注入原则

### 核心概念
- 使用**接口而非具体类型**进行依赖注入
- 遵循**依赖倒置原则**，高层模块不依赖低层模块
- 保持**Provider函数简洁**，单一职责
- 避免**循环依赖**，合理设计层次结构

### Wire项目结构
```
internal/
├── wire/
│   ├── wire.go          # Wire配置文件
│   └── wire_gen.go      # Wire生成的代码
├── handlers/            # HTTP处理器
├── services/            # 业务服务
├── repositories/        # 数据访问
└── infrastructure/      # 基础设施
```

## 📦 Provider设计模式

### 基础Provider示例
```go
// ✅ 好的Provider设计
func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func NewUserService(repo UserRepository, logger *zap.Logger) UserService {
    return &userService{
        repo:   repo,
        logger: logger,
    }
}

func NewUserHandler(service UserService, logger *zap.Logger) *UserHandler {
    return &UserHandler{
        service: service,
        logger:  logger,
    }
}
```

### Wire配置文件
```go
//go:build wireinject
// +build wireinject

package wire

import (
    "github.com/google/wire"
    "github.com/varluffy/shield/internal/handlers"
    "github.com/varluffy/shield/internal/services"
    "github.com/varluffy/shield/internal/repositories"
)

// 定义Provider集合
var repositorySet = wire.NewSet(
    repositories.NewUserRepository,
    repositories.NewPermissionRepository,
    repositories.NewRoleRepository,
)

var serviceSet = wire.NewSet(
    services.NewUserService,
    services.NewPermissionService,
    services.NewRoleService,
)

var handlerSet = wire.NewSet(
    handlers.NewUserHandler,
    handlers.NewPermissionHandler,
    handlers.NewRoleHandler,
)

// 应用程序初始化
func InitializeApp(configPath string) (*App, func(), error) {
    wire.Build(
        // 基础设施
        infrastructureSet,
        // 数据层
        repositorySet,
        // 业务层
        serviceSet,
        // 控制层
        handlerSet,
        // 应用程序
        NewApp,
    )
    return &App{}, nil, nil
}
```

## 🎯 分层架构集成

### 架构层次
```
┌─────────────────┐
│   HTTP Layer    │  ← Gin Handlers
├─────────────────┤
│  Business Layer │  ← Services
├─────────────────┤
│   Data Layer    │  ← Repositories
├─────────────────┤
│ Database Layer  │  ← GORM Models
└─────────────────┘
```

### 依赖注入流程
```go
// 1. 基础设施Provider
func NewDatabase(config *DatabaseConfig) (*gorm.DB, error) {
    // 数据库连接
}

func NewLogger(config *LogConfig) (*zap.Logger, error) {
    // 日志初始化
}

// 2. Repository层Provider
func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

// 3. Service层Provider
func NewUserService(
    repo UserRepository,
    logger *zap.Logger,
    txManager TransactionManager,
) UserService {
    return &userService{
        repo:      repo,
        logger:    logger,
        txManager: txManager,
    }
}

// 4. Handler层Provider
func NewUserHandler(
    service UserService,
    logger *zap.Logger,
) *UserHandler {
    return &UserHandler{
        service: service,
        logger:  logger,
    }
}
```

## 🔄 生命周期管理

### 应用程序启动
```go
func main() {
    // 1. 初始化应用
    app, cleanup, err := wire.InitializeApp("configs/config.dev.yaml")
    if err != nil {
        log.Fatalf("Failed to initialize app: %v", err)
    }
    defer cleanup()

    // 2. 设置路由
    router := setupRoutes(app)

    // 3. 启动服务器
    server := &http.Server{
        Addr:    ":8080",
        Handler: router,
    }

    // 4. 优雅关闭
    gracefulShutdown(server, cleanup)
}
```

### 资源清理
```go
func (app *App) Cleanup() {
    // 关闭数据库连接
    if sqlDB, err := app.DB.DB(); err == nil {
        sqlDB.Close()
    }
    
    // 关闭Redis连接
    if app.Redis != nil {
        app.Redis.Close()
    }
    
    // 清理其他资源
    app.Logger.Sync()
}
```

## 🚦 错误处理最佳实践

### 统一错误响应
```go
// 错误处理中间件
func ErrorHandlerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            
            switch e := err.Err.(type) {
            case *BusinessError:
                c.JSON(e.HTTPStatus(), gin.H{
                    "code":    e.Code,
                    "message": e.Message,
                    "trace_id": c.GetString("trace_id"),
                })
            default:
                c.JSON(http.StatusInternalServerError, gin.H{
                    "code":    500,
                    "message": "Internal Server Error",
                    "trace_id": c.GetString("trace_id"),
                })
            }
        }
    }
}
```

### Handler错误处理
```go
func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    user, err := h.service.GetUser(c.Request.Context(), userID)
    if err != nil {
        // 通过错误中间件处理
        c.Error(err)
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "code": 0,
        "data": user,
        "trace_id": c.GetString("trace_id"),
    })
}
```

## 📊 性能优化建议

### 路由优化
- 使用路由参数而非查询参数：`/users/:id` 而不是 `/users?id=123`
- 合理使用路由组，避免重复中间件执行
- 静态文件服务使用 `gin.Static()`

### 中间件优化
- 将重量级中间件放在路由组级别，而非全局
- 使用条件中间件，按需执行
- 缓存中间件结果，避免重复计算

### 依赖注入优化
- 使用单例模式管理重量级资源（数据库连接、Redis客户端）
- 延迟初始化非关键依赖
- 合理设计Provider的生命周期

## 🧪 测试最佳实践

### Handler测试
```go
func TestUserHandler_GetUser(t *testing.T) {
    // 1. 创建测试用的Service mock
    mockService := &MockUserService{}
    handler := NewUserHandler(mockService, logger)
    
    // 2. 创建测试路由
    router := gin.New()
    router.GET("/users/:id", handler.GetUser)
    
    // 3. 执行测试请求
    req := httptest.NewRequest("GET", "/users/123", nil)
    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)
    
    // 4. 验证结果
    assert.Equal(t, http.StatusOK, resp.Code)
}
```

### Wire测试
```go
func TestWireIntegration(t *testing.T) {
    // 使用测试配置
    app, cleanup, err := InitializeTestApp("configs/test.yaml")
    require.NoError(t, err)
    defer cleanup()
    
    // 验证依赖注入
    assert.NotNil(t, app.UserHandler)
    assert.NotNil(t, app.UserService)
}
```

## 📚 相关文档

- [数据库使用指南](go-database-guide.md) - GORM和Redis使用
- [可观测性指南](go-observability.md) - 日志和监控
- [开发规则约束](../DEVELOPMENT_RULES.md) - 架构约束
- [权限系统设计](../business/architecture/permission-system.md) - 权限中间件

## 🔗 外部资源

- [Gin官方文档](https://gin-gonic.com/)
- [Wire用户指南](https://github.com/google/wire)
- [Go Web开发最佳实践](https://golang.org/doc/effective_go.html)