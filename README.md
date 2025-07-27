# UltraFit - Go微服务开发框架

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/your-org/ultrafit)

UltraFit 是一个基于 Go 的高性能微服务开发框架，采用清洁架构设计，集成了现代化的技术栈和最佳实践。

## 🚀 特性

### 核心框架
- **Web框架**: Gin - 高性能HTTP路由
- **依赖注入**: Wire - 编译时依赖注入
- **数据库ORM**: GORM - 强大的ORM框架
- **配置管理**: Viper - 多源配置管理
- **日志系统**: Zap - 结构化高性能日志

### 功能特性
- ✅ **多租户权限管理** - 基于角色的访问控制
- ✅ **图形验证码** - 支持多种验证码类型
- ✅ **JWT认证** - 安全的身份验证
- ✅ **分布式存储** - Redis支持，自动降级
- ✅ **链路追踪** - OpenTelemetry + Jaeger
- ✅ **HTTP客户端** - 内置重试和链路追踪
- ✅ **统一响应** - 标准化API响应格式
- ✅ **事务管理** - 声明式事务处理

### 开发体验
- 🔧 **本地开发模式** - 无需Docker，支持多人协作
- 📊 **可观测性** - 完整的日志、指标、追踪
- 🧪 **测试覆盖** - 单元测试+集成测试
- 📚 **完整文档** - 详细的开发指南
- 🛡️ **架构保护** - 分层架构强制检查

## 📦 技术栈

| 组件 | 技术 | 版本 | 说明 |
|------|------|------|------|
| Web框架 | Gin | v1.9+ | HTTP路由和中间件 |
| 依赖注入 | Wire | v0.5+ | 编译时依赖注入 |
| 数据库 | GORM | v1.25+ | ORM框架 |
| 数据库 | MySQL | 8.0+ | 主数据库 |
| 缓存 | Redis | 6.0+ | 分布式缓存 |
| 配置 | Viper | v1.17+ | 配置管理 |
| 日志 | Zap | v1.26+ | 结构化日志 |
| 追踪 | OpenTelemetry | v1.21+ | 链路追踪 |
| 验证码 | base64Captcha | v1.2+ | 图形验证码 |
| 测试 | Testify | v1.8+ | 测试框架 |

## 🏗️ 架构设计

```
├── cmd/                    # 应用程序入口
│   ├── server/            # Web服务器
│   └── migrate/           # 数据库迁移
├── configs/               # 配置文件
├── internal/              # 内部应用代码
│   ├── handlers/          # HTTP处理器 (Controller层)
│   ├── services/          # 业务逻辑 (Service层)
│   ├── repositories/      # 数据访问 (Repository层)
│   ├── models/           # 数据模型
│   ├── dto/              # 数据传输对象
│   ├── config/           # 配置结构
│   ├── middleware/       # 中间件
│   └── wire/             # 依赖注入配置
├── pkg/                  # 可复用的包
│   ├── captcha/          # 验证码服务
│   ├── logger/           # 日志组件
│   ├── response/         # 响应格式
│   └── ...
└── docs/                 # 项目文档
```

## 🚀 快速开始

### 1. 环境要求

**必需**:
- Go 1.21+
- MySQL 8.0+

**可选**:
- Redis 6.0+ (验证码存储，可降级为内存)
- Jaeger (链路追踪，可禁用)

### 2. 安装项目

```bash
# 克隆项目
git clone https://github.com/your-org/ultrafit.git
cd ultrafit

# 安装工具
make tools

# 下载依赖
go mod download
```

### 3. 配置环境

```bash
# 复制配置文件 (可选)
cp configs/config.dev.yaml configs/config.local.yaml

# 编辑配置，修改数据库连接信息
vim configs/config.dev.yaml
```

### 4. 初始化数据库

```sql
-- 创建数据库
CREATE DATABASE ultrafit_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

```bash
# 运行数据库迁移
make migrate
```

### 5. 启动应用

```bash
# 快速启动（推荐）
make quick-start

# 或分步执行
make wire    # 生成依赖注入代码
make dev     # 启动开发服务器
```

### 6. 验证启动

```bash
# 健康检查
curl http://localhost:8080/health
# 返回: {"code":0,"message":"success","data":{"app":"ultrafit","status":"ok","version":"1.0.0"},"timestamp":"2024-01-01T12:00:00Z"}

# 测试验证码API
curl http://localhost:8080/api/v1/captcha/generate
# 返回: {"code":0,"message":"success","data":{"captcha_id":"xxx","captcha_image":"data:image/png;base64,xxx"},"timestamp":"2024-01-01T12:00:00Z"}
```

## 📋 常用命令

### 开发命令
```bash
make help           # 显示所有可用命令
make dev            # 开发模式启动
make run            # 运行应用
make build          # 构建应用
make test           # 运行测试
make test-coverage  # 生成测试覆盖率报告
```

### 代码质量
```bash
make format         # 格式化代码
make lint           # 代码检查
make full-check     # 完整检查（测试+lint）
make tidy           # 整理依赖
```

### 工具管理
```bash
make tools          # 安装开发工具
make check          # 检查开发环境
make wire           # 生成Wire代码
make migrate        # 数据库迁移
```

## 🔧 配置说明

### 基础配置

```yaml
# 应用配置
app:
  name: "ultrafit"
  version: "1.0.0"
  environment: "development"
  debug: true
  language: "zh"

# 服务器配置
server:
  host: "0.0.0.0"
  port: 8080

# 数据库配置
database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "your_password"
  name: "ultrafit_dev"
```

### 可选配置

```yaml
# Redis配置 (可选)
redis:
  # 启用Redis时取消注释
  # addrs: ["localhost:6379"]
  key_prefix: "ultrafit:"

# Jaeger配置 (可选)
jaeger:
  enabled: false
  otlp_url: "http://localhost:4318/v1/traces"

# 验证码配置
captcha:
  enabled: true
  type: "digit"
  width: 160
  height: 60
  length: 4
  expiration: 5m
```

## 📖 API文档

### 健康检查
```bash
GET /health
```

### 验证码API
```bash
# 生成验证码
GET /api/v1/captcha/generate

# 验证码校验
POST /api/v1/captcha/verify
Content-Type: application/json

{
  "captcha_id": "xxx",
  "captcha_code": "1234"
}
```

### 统一响应格式
```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## 🧪 测试

### 运行测试
```bash
# 所有测试
make test

# 测试覆盖率
make test-coverage

# 特定模块测试
go test -v ./pkg/captcha/...
```

### 测试示例
```bash
# 验证码功能测试
go test -v ./test/ -run TestCaptcha

# 完整API测试
go test -v ./test/ -run TestSimplifiedAPI
```

## 📚 项目文档

### 架构文档
- [核心架构](docs/architecture/go-microservices-core.md)
- [Wire依赖注入](docs/architecture/wire-architecture.md)

### 框架文档
- [Gin Web框架](docs/frameworks/go-gin-web.md)
- [GORM数据库](docs/frameworks/go-gorm-database.md)
- [可观测性](docs/frameworks/go-observability-logging.md)
- [配置管理](docs/frameworks/go-viper-config.md)

### 业务文档
- [系统设计](docs/business/SYSTEM_DESIGN_SUMMARY.md)
- [API文档](docs/business/api/)
- [数据库设计](docs/business/database/schema-design.md)

### 开发指南
- [开发环境配置](docs/DEVELOPMENT_SETUP.md)
- [开发规范](docs/DEVELOPMENT_RULES.md)
- [快速开始](docs/GETTING_STARTED.md)

## 🛡️ 开发规范

### 代码架构
- 严格遵循清洁架构分层
- Handler -> Service -> Repository
- 使用接口进行解耦
- 通过Wire进行依赖注入

### 开发流程
1. 需求分析和设计
2. 编写接口定义
3. 实现业务逻辑
4. 编写单元测试
5. 集成测试验证
6. 代码审查提交

### 质量保证
- 单元测试覆盖率 > 80%
- 代码格式化: `make format`
- 代码检查: `make lint`
- 完整检查: `make full-check`

## 🔍 常见问题

### 数据库连接失败
```bash
# 检查MySQL是否启动
mysql -u root -p

# 检查配置文件
vim configs/config.dev.yaml
```

### Redis连接失败
系统会自动降级为内存存储，这是正常的。

### 端口冲突
```yaml
# 修改端口
server:
  port: 8081
```

### 工具未安装
```bash
# 安装所有开发工具
make tools

# 检查环境
make check
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交改动 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证。详情请参考 [LICENSE](LICENSE) 文件。

## 📞 联系我们

- 项目地址: https://github.com/your-org/ultrafit
- 问题反馈: https://github.com/your-org/ultrafit/issues
- 技术讨论: [技术群/论坛链接]

---

⭐ 如果这个项目对你有帮助，请给我们一个星标！ 