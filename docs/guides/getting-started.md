# UltraFit 新人入门指南

欢迎加入UltraFit项目！本指南将帮助你快速上手，了解项目结构和开发流程。

## 🎯 项目概述

UltraFit是一个基于Go语言的微服务项目，采用清洁架构设计，集成了现代化的开发工具和最佳实践。

### 核心技术栈
- **Web框架**: Gin
- **依赖注入**: Wire
- **ORM**: GORM + MySQL
- **日志**: Zap + OpenTelemetry
- **配置**: Viper
- **测试**: Testify

### 项目特色
- 🏗️ 清洁架构：Handler → Service → Repository
- 🔍 完整可观测性：自动TraceID注入
- 📝 结构化日志：JSON格式，便于查询
- 🔧 自动化工具：Wire代码生成，依赖注入
- 🌐 多语言支持：中英文验证错误信息

## 📁 项目结构

```
ultrafit/
├── cmd/                    # 程序入口
│   ├── server/            # 主服务
│   └── migrate/           # 数据库迁移
├── internal/              # 核心业务逻辑（私有）
│   ├── handlers/          # HTTP处理器
│   ├── services/          # 业务逻辑服务
│   ├── repositories/      # 数据访问层
│   ├── models/            # 数据模型
│   ├── dto/               # 数据传输对象
│   ├── config/            # 配置管理
│   └── wire/              # 依赖注入配置
├── pkg/                   # 公共工具包
│   ├── logger/            # 日志工具
│   ├── response/          # 响应工具
│   ├── validator/         # 验证工具
│   └── tracing/           # 追踪工具
├── configs/               # 配置文件
├── test/                  # 测试文件
├── docs/                  # 文档
└── scripts/               # 脚本工具
```

## 🚀 快速开始

### 1. 环境准备
```bash
# 确保Go版本 >= 1.21
go version

# 克隆项目
git clone <repository-url>
cd ultrafit

# 安装依赖
go mod tidy
```

### 2. 安装开发工具
```bash
# 一键安装所有必要工具
make install-tools

# 或手动安装
go install github.com/google/wire/cmd/wire@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 3. 启动项目
```bash
# 完整开发环境启动（推荐）
make dev

# 这个命令会自动：
# - 生成Wire代码
# - 启动MySQL（如果需要）
# - 运行数据库迁移
# - 启动应用服务器
```

### 4. 验证启动
```bash
# 健康检查
curl http://localhost:8080/health

# 创建测试用户
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"测试用户","email":"test@example.com","password":"password123"}'
```

## 🏗️ 架构理解

### 清洁架构分层

```
┌─────────────────┐
│   Handler层      │  HTTP处理、参数验证、响应格式化
├─────────────────┤
│   Service层      │  业务逻辑、事务管理、编排调用
├─────────────────┤
│  Repository层    │  数据访问、SQL操作、缓存管理
├─────────────────┤
│   Model层        │  数据模型、业务实体
└─────────────────┘
```

### 依赖方向（重要！）
- Handler → Service ✅
- Service → Repository ✅
- Repository → Database ✅
- Handler → Repository ❌（禁止跨层调用）

### 接口驱动开发
```go
// 定义接口
type UserService interface {
    CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
}

// 实现接口
type UserServiceImpl struct {
    userRepo repositories.UserRepository
}

// Wire绑定
wire.Bind(new(UserService), new(*UserServiceImpl))
```

## 📝 开发规范

### 1. 新增功能流程
```
1. 在dto/中定义请求/响应结构
2. 在models/中定义数据模型
3. 在repositories/中实现数据访问
4. 在services/中实现业务逻辑
5. 在handlers/中实现HTTP处理
6. 在wire/中配置依赖注入
7. 运行make wire生成代码
```

### 2. 代码规范
- **函数命名**: 驼峰命名，动词开头
- **接口定义**: 以业务含义命名，不用I前缀
- **错误处理**: 使用pkg/errors包装错误
- **日志记录**: 使用结构化日志，自动包含TraceID
- **测试覆盖**: 核心业务逻辑必须有测试

### 3. 提交规范
```bash
# 生成代码
make wire

# 格式化代码
make format

# 运行测试
make test

# 代码检查
make lint
```

## 🔧 常用命令

### 开发命令
```bash
make dev           # 开发模式启动
make build         # 构建应用
make test          # 运行测试
make wire          # 生成Wire代码
make format        # 格式化代码
make lint          # 代码检查
```

### 数据库命令
```bash
make migrate       # 运行数据库迁移
make dev-db        # 启动开发数据库
```

### 工具命令
```bash
make install-tools # 安装开发工具
make clean         # 清理构建文件
make help          # 查看所有命令
```

## 🐛 常见问题

### Q1: Wire生成失败
```bash
# 确保安装了Wire
go install github.com/google/wire/cmd/wire@latest

# 重新生成
make wire
```

### Q2: 数据库连接失败
```bash
# 检查配置文件 configs/config.dev.yaml
# 确保MySQL已启动
make dev-db
```

### Q3: 日志中没有TraceID
```bash
# 确保使用了正确的logger
logger.InfoWithTrace(ctx, "message", zap.String("key", "value"))
```

### Q4: 测试失败
```bash
# 检查测试配置
# 确保测试数据库已启动
make test
```

## 📚 深入学习

### 核心文档（按优先级）
1. **[架构设计](architecture/go-microservices-core.md)** - 理解整体架构
2. **[Web开发](frameworks/go-gin-web.md)** - Gin框架使用
3. **[依赖注入](frameworks/go-wire-di.md)** - Wire使用指南
4. **[数据库操作](frameworks/go-gorm-database.md)** - GORM使用
5. **[日志追踪](frameworks/go-observability-logging.md)** - 可观测性
6. **[配置管理](frameworks/go-viper-config.md)** - 配置使用

### 代码示例
```go
// 完整的功能实现示例
// 1. DTO定义
type CreateUserRequest struct {
    Name     string `json:"name" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

// 2. Handler实现
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.responseWriter.Error(c, err)
        return
    }
    
    resp, err := h.userService.CreateUser(c.Request.Context(), &req)
    if err != nil {
        h.responseWriter.Error(c, err)
        return
    }
    
    h.responseWriter.Success(c, resp)
}

// 3. Service实现
func (s *UserServiceImpl) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
    // 业务逻辑
    user := &models.User{
        Name:  req.Name,
        Email: req.Email,
        // ...
    }
    
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    return &dto.UserResponse{
        ID:    user.ID,
        Name:  user.Name,
        Email: user.Email,
    }, nil
}
```

## 🎯 下一步

1. **熟悉项目结构**：了解各目录的作用
2. **阅读核心文档**：理解架构设计原理
3. **运行示例代码**：体验开发流程
4. **实践小功能**：尝试添加新的API
5. **参与代码审查**：学习最佳实践

## 💡 开发建议

- **先理解架构**：不要急于写代码，先理解分层设计
- **遵循接口**：始终通过接口而非具体实现进行调用
- **重视测试**：为核心业务逻辑编写测试
- **关注日志**：善用结构化日志排查问题
- **保持简洁**：避免过度设计，专注业务价值

---

🎉 **欢迎加入UltraFit团队！有问题随时交流。** 