# UltraFit 开发环境配置指南

## 📖 概述

UltraFit 采用**本地环境开发模式**，不包含 Docker Compose 等基础设施配置。开发者需要使用自己本地的 MySQL、Redis、Jaeger 等服务。

## 🔧 环境要求

### 必需服务
- **MySQL 8.0+**: 主数据库
- **Go 1.21+**: 运行环境

### 可选服务  
- **Redis**: 验证码存储（可自动降级为内存存储）
- **Jaeger**: 链路追踪（可禁用）

## ⚙️ 配置说明

### 1. 基础配置文件

项目使用 `configs/config.dev.yaml` 作为开发环境的参考配置。每个开发者可以根据自己的环境进行调整。

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
  host: "0.0.0.0"  # 监听地址
  port: 8080       # 监听端口

# 数据库配置 (必需)
database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "123456"  # 修改为你的MySQL密码
  name: "ultrafit_dev"

# Redis配置 (可选)
redis:
  # 如果有Redis，取消注释下面一行
  # addrs: ["localhost:6379"]
  key_prefix: "ultrafit:"

# JWT认证配置
auth:
  jwt:
    secret: "ultrafit-dev-secret-key-change-in-production"
    expires_in: 24h
    issuer: "ultrafit"

# HTTP客户端配置
http_client:
  timeout: 30
  retry_count: 3
  enable_trace: true

# Jaeger配置 (可选)
jaeger:
  enabled: false  # 如需启用链路追踪，设置为true
  otlp_url: "http://localhost:4318/v1/traces"

# 验证码配置
captcha:
  enabled: true
  type: "digit"
  width: 160
  height: 60
  length: 4
  noise_count: 5
  expiration: 5m
```

### 2. 数据库配置

#### MySQL 数据库创建
```sql
-- 创建数据库
CREATE DATABASE ultrafit_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户 (可选，也可以使用root)
CREATE USER 'ultrafit'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON ultrafit_dev.* TO 'ultrafit'@'localhost';
FLUSH PRIVILEGES;
```

#### 数据库操作规范

**重要**: 本项目使用 **MCP (Model Context Protocol)** 进行数据库操作，请遵循以下规范：

1. **禁止直接命令行操作**: 不要使用 `mysql` 命令行客户端直接操作数据库
2. **仅操作指定数据库**: 只能操作 `ultrafit_dev` 数据库，禁止操作其他数据库
3. **使用MCP工具**: 所有数据库查询、更新、结构变更都通过MCP工具进行
4. **安全第一**: MCP工具提供更好的安全性和操作记录

**示例操作**:
```sql
-- 查看表结构
SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'ultrafit_dev';

-- 查询数据
SELECT * FROM users LIMIT 10;

-- 插入数据
INSERT INTO users (username, email) VALUES ('test', 'test@example.com');
```

#### 常见MySQL配置
```yaml
# 本地MySQL (默认)
database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "your_mysql_password"

# Docker MySQL
database:
  host: "localhost"
  port: 3306  # 或你的映射端口
  user: "root" 
  password: "your_password"

# 远程MySQL
database:
  host: "192.168.1.100"
  port: 3306
  user: "ultrafit"
  password: "your_password"
```

### 3. Redis配置 (可选)

Redis主要用于验证码存储。如果没有Redis，系统会自动降级为内存存储。

```yaml
# 启用Redis
redis:
  addrs: ["localhost:6379"]
  password: ""
  db: 0

# 禁用Redis (使用内存存储)
redis:
  # addrs: []  # 注释掉或留空
```

### 4. Jaeger配置 (可选)

```yaml
# 启用Jaeger
jaeger:
  enabled: true
  otlp_url: "http://localhost:4318/v1/traces"

# 禁用Jaeger
jaeger:
  enabled: false
```

## 🚀 快速启动

### 1. 克隆项目
```bash
git clone <repository-url>
cd ultrafit
```

### 2. 安装工具和依赖
```bash
# 安装开发工具
make tools

# 下载依赖
go mod download
```

### 3. 配置环境
```bash
# 复制配置文件 (可选)
cp configs/config.dev.yaml configs/config.local.yaml

# 编辑配置文件，修改数据库等信息
vim configs/config.dev.yaml
# 或
vim configs/config.local.yaml
```

### 4. 初始化数据库
```bash
# 运行数据库迁移
make migrate
```

### 5. 启动应用
```bash
# 快速启动（包含工具安装、代码生成、迁移）
make quick-start

# 或分步执行
make wire    # 生成依赖注入代码
make dev     # 启动开发服务器

# 或直接运行
make run
```

### 6. 验证启动
```bash
# 健康检查
curl http://localhost:8080/health

# 测试验证码API
curl http://localhost:8080/api/v1/captcha/generate
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

### 生产构建
```bash
make build-prod     # 生产环境构建
```

## 🔧 常见问题

### 1. 数据库连接失败
```
Error 1045 (28000): Access denied for user 'root'@'localhost'
```

**解决方案**: 
- 检查MySQL密码是否正确
- 确认MySQL服务是否启动
- 检查用户权限

### 2. Redis连接失败
如果Redis不可用，系统会自动使用内存存储，这是正常的。

### 3. 端口冲突
```
bind: address already in use
```

**解决方案**: 修改配置文件中的端口
```yaml
server:
  port: 8081  # 改为其他端口
```

### 4. Wire代码生成失败
```bash
# 重新安装wire工具
make tools

# 手动生成代码
go generate ./internal/wire/...
```

### 5. 工具未安装
```bash
# 安装所有开发工具
make tools

# 检查环境
make check
```

## 🌟 开发提示

### 1. 配置文件优先级
1. 命令行指定的配置文件 (`-config` 参数)
2. `configs/config.dev.yaml` (默认)

### 2. 环境变量覆盖
可以使用环境变量覆盖配置：
```bash
export ULTRAFIT_DB_PASSWORD=your_password
export ULTRAFIT_SERVER_PORT=8081
```

### 3. 本地配置文件
建议创建 `configs/config.local.yaml` 用于本地开发，该文件已被 `.gitignore` 忽略。

### 4. 多人协作
- 不要提交个人的数据库密码等敏感信息
- `configs/config.dev.yaml` 保持通用配置
- 使用 `configs/config.local.yaml` 存储个人配置

### 5. 测试覆盖率
```bash
make test-coverage
# 生成的 coverage.html 可在浏览器中查看
```

## 📚 相关文档

- [项目架构文档](./architecture/go-microservices-core.md)
- [API文档](./business/api/)
- [数据库设计](./business/database/schema-design.md)

## 🆘 技术支持

如遇到配置问题，可以：
1. 检查日志文件 `logs/app.log`
2. 运行 `make check` 检查环境
3. 参考项目文档
4. 联系项目维护者 