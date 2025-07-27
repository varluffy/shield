# UltraFit 快速开始指南

欢迎加入 UltraFit 项目！本指南将帮助你在 15 分钟内完成环境搭建并运行项目。

## 🎯 项目概述

UltraFit 是一个基于 Go 的微服务开发框架，实现了清洁架构和多租户权限管理系统。

### 核心技术栈
- **Web框架**: Gin - 高性能 HTTP 框架
- **依赖注入**: Wire - Google 官方 DI 工具
- **ORM**: GORM + MySQL - 数据库操作
- **日志**: Zap + OpenTelemetry - 结构化日志和追踪
- **配置**: Viper - 配置管理
- **测试**: Testify + Gomock - 单元测试和集成测试

### 项目特色
- 🏗️ **清洁架构**: Handler → Service → Repository 严格分层
- 🔍 **完整可观测性**: 自动 TraceID 注入，分布式追踪
- 📝 **结构化日志**: JSON 格式，支持日志聚合分析
- 🔧 **自动化工具**: Wire 代码生成，依赖注入自动化
- 🌐 **多租户支持**: 完整的租户隔离和权限管理
- 🛡️ **安全机制**: JWT 认证、图形验证码、权限控制

## 📁 项目结构

```
shield/
├── cmd/                    # 应用程序入口
│   ├── server/            # 主 Web 服务器
│   ├── migrate/           # 数据库迁移工具
│   └── admin/             # 管理工具
├── internal/              # 核心业务逻辑（私有）
│   ├── handlers/          # HTTP 请求处理器
│   ├── services/          # 业务逻辑服务
│   ├── repositories/      # 数据访问层
│   ├── models/            # 数据模型定义
│   ├── dto/               # 数据传输对象
│   ├── middleware/        # HTTP 中间件
│   ├── config/            # 配置管理
│   └── wire/              # 依赖注入配置
├── pkg/                   # 可复用的公共包
│   ├── auth/              # JWT 认证工具
│   ├── captcha/           # 验证码服务
│   ├── logger/            # 日志工具
│   ├── response/          # HTTP 响应工具
│   └── errors/            # 错误处理
├── configs/               # 配置文件
├── test/                  # 集成测试
├── docs/                  # 项目文档
└── scripts/               # 工具脚本
```

## 🚀 快速开始

### 1. 环境要求

**必需服务**:
- **Go 1.21+**: 主要运行环境
- **MySQL 8.0+**: 主数据库

**可选服务**:
- **Redis**: 验证码存储（自动降级为内存存储）
- **Jaeger**: 分布式追踪（可禁用）

### 2. 项目初始化

```bash
# 克隆项目
git clone <repository-url>
cd shield

# 一键初始化（推荐）
make init              # 完整初始化（包含管理员创建）
# 或
make quick-init        # 快速初始化（仅开发环境）

# 验证启动
curl http://localhost:8080/health
```

### 3. 手动步骤（可选）

如果一键初始化失败，可以手动执行以下步骤：

```bash
# 安装开发工具
make deps

# 配置数据库
mysql -u root -p
CREATE DATABASE shield CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 运行数据库迁移
make migrate

# 生成 Wire 代码
make wire

# 启动开发服务器
make run
```

### 4. 验证安装

```bash
# 健康检查
curl http://localhost:8080/health

# 测试验证码生成
curl http://localhost:8080/api/v1/captcha/generate

# 测试登录 API（开发环境）
curl -X POST http://localhost:8080/api/v1/auth/test-login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123","tenant_id":"1"}'
```

## ⚙️ 配置说明

### 配置文件优先级
1. `configs/config.local.yaml` - 本地覆盖配置（Git 忽略）
2. `configs/config.dev.yaml` - 开发环境默认配置

### 基础配置示例

```yaml
# configs/config.local.yaml（推荐用于本地开发）
database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "your_mysql_password"    # 修改为你的密码
  name: "shield"

# Redis 配置（可选）
redis:
  addrs: ["localhost:6379"]          # 有 Redis 时启用
  # addrs: []                        # 无 Redis 时注释或留空

# JWT 认证
auth:
  jwt:
    secret: "your-secret-key"        # 生产环境请使用随机密钥
    expires_in: 24h
```

## 📋 常用开发命令

### 服务管理
```bash
make run               # 启动开发服务器（自动停止旧服务）
make safe-run          # 安全启动（不杀死已有进程）
make stop-service      # 停止所有 shield 服务
make status           # 检查服务状态
make restart-service   # 重启服务
```

### 开发工作流
```bash
make wire             # 重新生成依赖注入代码
make test             # 运行所有测试
make docs             # 生成 API 文档
bash scripts/quality-check.sh  # 完整质量检查
```

### 数据库操作
```bash
make migrate          # 运行数据库迁移
make create-admin     # 创建管理员用户

# 数据库查询（使用 MCP 工具）
# 查看所有表
SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'shield';
```

## 🏗️ 架构快速理解

### 清洁架构分层
```
HTTP Request
    ↓
Handler Layer    ← 参数绑定、响应格式化
    ↓
Service Layer    ← 业务逻辑、事务管理
    ↓
Repository Layer ← 数据访问、数据库操作
    ↓
Database
```

### 关键原则
- **单向依赖**: Handler → Service → Repository
- **接口驱动**: 所有跨层调用通过接口
- **依赖注入**: 使用 Wire 自动生成 DI 代码
- **上下文传播**: Context 贯穿整个请求生命周期

### 多租户架构
- **租户隔离**: 所有用户数据包含 `tenant_id`
- **权限控制**: 菜单、按钮、API、字段四级权限
- **JWT 上下文**: Token 包含租户信息

## 🔧 常见问题

### 端口被占用
```bash
make kill-port        # 杀死占用 8080 端口的进程
# 或修改配置文件中的端口
```

### Wire 代码生成失败
```bash
make wire             # 重新生成依赖注入代码
```

### 数据库连接失败
- 检查 MySQL 服务是否启动
- 确认数据库密码配置正确
- 确保数据库 `shield` 已创建

### Redis 连接失败
系统会自动降级为内存存储，这是正常行为。

## 🌟 开发提示

### 新功能开发流程
1. **创建模型**: `internal/models/` 中定义数据模型
2. **实现 Repository**: `internal/repositories/` 中实现数据访问
3. **实现 Service**: `internal/services/` 中实现业务逻辑
4. **实现 Handler**: `internal/handlers/` 中实现 HTTP 处理
5. **添加路由**: 在 `internal/routes/routes.go` 中注册路由
6. **重新生成**: 运行 `make wire` 重新生成依赖注入代码

### 测试策略
```bash
# 运行特定测试
go test -v ./test/ -run TestCaptcha
go test -v ./test/ -run TestPermission

# 生成覆盖率报告
go test -v -cover ./...
```

### 调试技巧
- 查看日志中的 `trace_id` 追踪请求流程
- 使用 `make status` 检查服务状态
- 后台服务日志保存在 `app.log` 文件中

## 📚 下一步

- 📖 [架构详细说明](./architecture.md) - 深入了解系统架构
- 🔧 [API 开发指南](./api-guide.md) - 学习 API 开发规范
- 🧪 [测试指南](./testing-guide.md) - 掌握测试最佳实践

## 🆘 获取帮助

如遇到问题，可以：
1. 查看详细错误日志
2. 运行 `bash scripts/quality-check.sh` 检查环境
3. 参考项目文档 `docs/` 目录
4. 联系项目维护者