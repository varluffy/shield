# UltraFit 系统使用指南

本文档整合了登录系统和服务管理的使用指南，帮助开发者快速掌握系统操作。

## 🔐 登录系统使用

### 系统架构

#### 核心组件
1. **AuthMiddleware**: 认证中间件，负责JWT令牌验证
2. **UserService**: 用户服务层，处理登录业务逻辑
3. **JWT服务**: 生成和验证JWT令牌
4. **bcrypt密码加密**: 使用bcrypt算法加密用户密码

#### 登录流程
```mermaid
sequenceDiagram
    participant C as Client
    participant H as Handler
    participant S as Service
    participant R as Repository
    participant DB as Database
    participant JWT as JWT Service

    C->>H: POST /auth/login
    H->>H: 绑定请求参数
    H->>S: Login()
    S->>S: 验证验证码
    S->>R: GetByEmail()
    R->>DB: 查询用户
    DB-->>R: 返回用户信息
    R-->>S: 返回用户模型
    S->>S: 验证密码(bcrypt)
    S->>JWT: 生成JWT令牌
    JWT-->>S: 返回令牌
    S-->>H: 返回登录响应
    H-->>C: 返回JWT令牌
```

### 测试账号

#### 预设测试账号
| 邮箱 | 密码 | 角色 | 租户ID | 状态 |
|------|------|------|--------|------|
| admin@example.com | admin123 | 系统管理员 | 1 | active |

#### 密码加密说明
```go
// 密码使用bcrypt加密，成本因子为10
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
```

### API使用示例

#### 1. 生成验证码
```bash
curl -X GET "http://localhost:8080/api/v1/captcha/generate"
```

响应：
```json
{
  "code": 0,
  "message": "验证码生成成功",
  "data": {
    "captcha_id": "bp8RkzOTBEObGLvueygk",
    "captcha_image": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAJYAAAA..."
  }
}
```

#### 2. 用户登录
```bash
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123",
    "captcha_id": "bp8RkzOTBEObGLvueygk",
    "captcha_answer": "8849"
  }'
```

响应：
```json
{
  "code": 0,
  "message": "登录成功",
  "data": {
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "tenant_id": 0,
      "is_system_admin": true
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIs...",
      "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g...",
      "token_type": "Bearer",
      "expires_in": 7200
    }
  }
}
```

#### 3. 测试登录（无验证码）
```bash
curl -X POST "http://localhost:8080/api/v1/auth/test-login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123"
  }'
```

#### 4. 刷新令牌
```bash
curl -X POST "http://localhost:8080/api/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g..."
  }'
```

#### 5. 获取用户权限
```bash
curl -X GET "http://localhost:8080/api/v1/user/permissions" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### 常见问题解决

#### 登录失败
```json
{
  "code": 2005,
  "message": "凭据无效"
}
```
**解决方案**：检查邮箱和密码是否正确

#### 验证码错误
```json
{
  "code": 2011,
  "message": "验证码错误"
}
```
**解决方案**：重新获取验证码

#### 令牌过期
```json
{
  "code": 1003,
  "message": "未授权"
}
```
**解决方案**：使用refresh_token刷新令牌

## 🛠️ 服务管理规范

### 问题背景

在开发过程中，经常遇到端口占用的问题：
- 服务启动失败：`bind: address already in use`
- 忘记停止旧服务，导致新服务无法启动
- 需要手动查找和杀死进程，效率低下

### 解决方案

#### 核心服务管理命令
```bash
# 检查端口状态
make check-port      # 检查8080端口是否被占用

# 服务停止
make stop-service    # 停止所有ultrafit相关服务
make kill-port       # 杀死占用8080端口的进程

# 服务启动
make run             # 开发模式启动（自动停止旧服务）
make safe-run        # 安全启动（先检查端口）
make start-service   # 后台启动服务

# 服务状态
make status          # 检查服务状态
make restart-service # 重启服务
```

#### 开发流程命令
```bash
# 推荐的开发流程
make init            # 首次项目初始化
make run             # 开发运行（自动处理端口占用）
make stop-service    # 停止服务
make status          # 检查状态
```

### 开发规范

#### ✅ 正确的做法

1. **使用规范命令启动服务**
   ```bash
   # 推荐：使用make命令
   make run
   
   # 或者安全启动
   make safe-run
   ```

2. **及时停止服务**
   ```bash
   # 开发结束后停止服务
   make stop-service
   
   # 或者使用Ctrl+C停止前台服务
   ```

3. **检查服务状态**
   ```bash
   # 检查端口占用
   make check-port
   
   # 检查服务状态
   make status
   ```

#### ❌ 错误的做法

1. **直接使用go run启动**
   ```bash
   # 不推荐：直接启动可能遇到端口占用
   go run cmd/server/main.go
   ```

2. **忘记停止旧服务**
   ```bash
   # 错误：不停止旧服务直接启动新服务
   make run  # 第一次启动
   make run  # 第二次启动 - 会失败
   ```

3. **手动杀进程**
   ```bash
   # 不推荐：手动查找和杀死进程
   ps aux | grep ultrafit
   kill -9 <pid>
   ```

### 故障排除

#### 端口占用问题
```bash
# 问题：端口8080被占用
Error: listen tcp :8080: bind: address already in use

# 解决方案1：使用make命令自动处理
make kill-port
make run

# 解决方案2：检查并手动处理
make check-port
# 如果有进程占用，会显示进程信息
make stop-service
```

#### 服务无响应
```bash
# 检查服务状态
make status

# 如果服务卡住，强制重启
make restart-service

# 查看日志排查问题
make logs
```

#### 数据库连接问题
```bash
# 检查数据库配置
cat configs/config.dev.yaml

# 测试数据库连接
make test-db

# 重新初始化数据库
make migrate
```

### 生产环境部署

#### 部署命令
```bash
# 构建生产版本
make build

# 启动生产服务
make start-prod

# 检查生产服务状态
make status-prod

# 停止生产服务
make stop-prod
```

#### 服务监控
```bash
# 健康检查
curl http://localhost:8080/health

# 查看服务指标
curl http://localhost:8080/metrics

# 查看应用日志
tail -f logs/ultrafit.log
```

### 开发最佳实践

#### 1. 开发环境启动流程
```bash
# 1. 首次启动
make init           # 安装依赖、初始化数据库

# 2. 日常开发
make run            # 启动开发服务器

# 3. 代码变更后
Ctrl+C             # 停止服务
make run           # 重新启动

# 4. 开发结束
make stop-service  # 确保服务完全停止
```

#### 2. 团队协作规范
```bash
# 代码提交前
make test          # 运行测试
make lint          # 代码检查
make format        # 代码格式化

# 代码拉取后
make tidy          # 整理依赖
make wire          # 重新生成依赖注入代码
make migrate       # 应用数据库迁移
```

#### 3. 问题调试流程
```bash
# 1. 检查服务状态
make status

# 2. 查看实时日志
make logs

# 3. 检查配置
make check-config

# 4. 测试API
make test-api

# 5. 重启服务
make restart-service
```

## 📚 相关文档

- [权限系统快速上手](permission-quick-start.md) - 权限系统使用指南
- [开发环境配置](../DEVELOPMENT_SETUP.md) - 环境配置详情
- [开发规则约束](../DEVELOPMENT_RULES.md) - 开发规范
- [API文档](../business/api/auth-api.md) - 认证API详细文档

## 🔗 快速链接

- [健康检查](http://localhost:8080/health) - 服务健康状态
- [API文档](http://localhost:8080/swagger/) - Swagger API文档
- [Makefile命令](../../Makefile) - 所有可用命令