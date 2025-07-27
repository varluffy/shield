# Wire 模块化依赖注入架构

## 📋 目录结构

本项目采用模块化的Wire依赖注入设计，将Provider按层级和模块进行组织，避免多人开发时的代码冲突。

```
internal/
├── infrastructure/
│   └── providers.go          # 基础设施层ProviderSet (配置、日志、数据库、追踪)
├── repositories/
│   └── providers.go          # Repository层ProviderSet (数据访问层)
├── services/
│   └── providers.go          # Service层ProviderSet (业务逻辑层)
├── handlers/
│   └── providers.go          # Handler层ProviderSet (HTTP处理层)
└── wire/
    ├── wire.go               # 主Wire配置文件
    └── wire_gen.go           # Wire自动生成文件
```

## 🏗️ 架构设计原则

### 1. 模块化分离
- **基础设施层**: 配置、日志、数据库、追踪等基础组件
- **Repository层**: 数据访问对象，负责与数据库交互
- **Service层**: 业务逻辑处理，调用Repository层
- **Handler层**: HTTP请求处理，调用Service层

### 2. 避免冲突
- 每个模块维护自己的ProviderSet
- 多人开发时只需要修改对应模块的providers.go
- 主wire.go只负责组合各模块的ProviderSet

### 3. 清晰的依赖关系
```
Handler -> Service -> Repository -> Database
   ↓         ↓          ↓           ↓
ResponseWriter -> Logger -> Config
```

## 📁 模块详解

### Infrastructure Layer (基础设施层)
**文件**: `internal/infrastructure/providers.go`

```go
var ProviderSet = wire.NewSet(
    ProvideConfig,      // 配置提供者
    ProvideLogger,      // 日志提供者
    ProvideTracer,      // 追踪提供者
    ProvideDatabase,    // 数据库提供者
)
```

**职责**:
- 系统配置管理
- 日志系统初始化
- 数据库连接管理
- 分布式追踪配置

### Repository Layer (数据访问层)
**文件**: `internal/repositories/providers.go`

```go
var ProviderSet = wire.NewSet(
    NewUserRepository,
    // NewProductRepository,    // 示例：产品Repository
    // NewOrderRepository,      // 示例：订单Repository
)
```

**职责**:
- 数据库操作封装
- 数据访问接口实现
- 查询优化和缓存

### Service Layer (业务逻辑层)
**文件**: `internal/services/providers.go`

```go
var ProviderSet = wire.NewSet(
    NewUserService,
    // NewProductService,       // 示例：产品Service
    // NewOrderService,         // 示例：订单Service
)
```

**职责**:
- 业务逻辑处理
- 事务管理
- 业务规则验证

### Handler Layer (HTTP处理层)
**文件**: `internal/handlers/providers.go`

```go
var ProviderSet = wire.NewSet(
    response.NewResponseWriter,  // 响应处理器
    NewUserHandler,
    // NewProductHandler,       // 示例：产品Handler
    // NewOrderHandler,         // 示例：订单Handler
)
```

**职责**:
- HTTP请求处理
- 参数验证和绑定
- 响应格式化

## 🔧 使用方法

### 1. 添加新的Repository

```go
// 1. 在 internal/repositories/ 目录下创建新的Repository
type ProductRepository struct {
    db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
    return &ProductRepository{db: db}
}

// 2. 在 internal/repositories/providers.go 中添加
var ProviderSet = wire.NewSet(
    NewUserRepository,
    NewProductRepository,    // 添加这行
)
```

### 2. 添加新的Service

```go
// 1. 在 internal/services/ 目录下创建新的Service
type ProductService struct {
    repo *repositories.ProductRepository
    logger *logger.Logger
}

func NewProductService(
    repo *repositories.ProductRepository,
    logger *logger.Logger,
) *ProductService {
    return &ProductService{
        repo: repo,
        logger: logger,
    }
}

// 2. 在 internal/services/providers.go 中添加
var ProviderSet = wire.NewSet(
    NewUserService,
    NewProductService,       // 添加这行
)
```

### 3. 添加新的Handler

```go
// 1. 在 internal/handlers/ 目录下创建新的Handler
type ProductHandler struct {
    service *services.ProductService
    responseWriter *response.ResponseWriter
}

func NewProductHandler(
    service *services.ProductService,
    responseWriter *response.ResponseWriter,
) *ProductHandler {
    return &ProductHandler{
        service: service,
        responseWriter: responseWriter,
    }
}

// 2. 在 internal/handlers/providers.go 中添加
var ProviderSet = wire.NewSet(
    response.NewResponseWriter,
    NewUserHandler,
    NewProductHandler,       // 添加这行
)
```

### 4. 更新应用程序结构

```go
// 在 internal/wire/wire.go 中更新App结构
type App struct {
    Config         *config.Config
    Logger         *logger.Logger
    DB             *gorm.DB
    UserHandler    *handlers.UserHandler
    ProductHandler *handlers.ProductHandler  // 添加新Handler
}

func NewApp(
    cfg *config.Config,
    logger *logger.Logger,
    db *gorm.DB,
    userHandler *handlers.UserHandler,
    productHandler *handlers.ProductHandler,  // 添加参数
) *App {
    return &App{
        Config:         cfg,
        Logger:         logger,
        DB:             db,
        UserHandler:    userHandler,
        ProductHandler: productHandler,        // 赋值
    }
}
```

## 🚀 代码生成

```bash
# 重新生成Wire代码
go generate ./...

# 编译项目
go build ./cmd/server
```

## ✅ 优势

### 1. 团队协作友好
- **避免冲突**: 每个开发者只需要修改自己负责模块的providers.go
- **清晰职责**: 按层级分离，职责明确
- **独立开发**: 模块间松耦合，可以并行开发

### 2. 维护性强
- **模块化管理**: 每个层级的依赖注入独立管理
- **易于扩展**: 添加新组件只需要在对应模块中添加
- **便于测试**: 每个模块可以独立进行单元测试

### 3. 可读性好
- **结构清晰**: 一目了然的模块组织
- **文档完善**: 每个ProviderSet都有清晰的注释说明
- **示例丰富**: 提供了完整的使用示例

## 🔗 相关文档

- [Go Wire 官方文档](https://github.com/google/wire)
- [项目整体架构文档](../go-microservices-core.md)
- [依赖注入最佳实践](../go-wire-di.md) 