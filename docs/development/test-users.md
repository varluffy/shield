# 测试用户管理指南

本文档介绍如何管理Shield项目的标准测试用户，包括创建、使用和维护测试用户账号。

## 🎯 概述

为了解决开发和测试过程中反复调试登录凭据的问题，Shield项目建立了标准测试用户系统。所有测试用户都有已知的密码，可以直接用于开发、测试和API调试。

## 👥 标准测试用户

系统预定义了以下标准测试用户：

### 系统管理员 (System Admin)
- **邮箱**: `admin@system.test`
- **密码**: `admin123`
- **租户ID**: `0` (系统租户)
- **角色**: `system_admin`
- **权限**: 系统全部权限（绕过所有权限检查）

### 租户管理员 (Tenant Admin)
- **邮箱**: `admin@tenant.test`
- **密码**: `admin123`
- **租户ID**: `1` (默认租户)
- **角色**: `tenant_admin`
- **权限**: 租户管理权限

### 普通用户 (Regular User)
- **邮箱**: `user@tenant.test`
- **密码**: `user123`
- **租户ID**: `1` (默认租户)
- **角色**: `user` (如果角色存在)
- **权限**: 基础用户权限

### 测试用户 (Test User)
- **邮箱**: `test@example.com`
- **密码**: `test123`
- **租户ID**: `1` (默认租户)
- **角色**: `user` (如果角色存在)
- **权限**: 基础用户权限

## 🛠️ 管理命令

使用以下命令管理测试用户：

### 创建标准测试用户
```bash
go run cmd/migrate/*.go -action=create-test-users -config=configs/config.dev.yaml
```

### 清理测试用户
```bash
go run cmd/migrate/*.go -action=clean-test-users -config=configs/config.dev.yaml
```

### 列出测试用户状态
```bash
go run cmd/migrate/*.go -action=list-test-users -config=configs/config.dev.yaml
```

## 🔐 登录测试

### 使用curl测试登录

#### 系统管理员登录
```bash
curl -X POST "http://localhost:8080/api/v1/auth/test-login" \\
  -H "Content-Type: application/json" \\
  -d '{
    "email": "admin@system.test",
    "password": "admin123",
    "tenant_id": "0"
  }'
```

#### 租户管理员登录
```bash
curl -X POST "http://localhost:8080/api/v1/auth/test-login" \\
  -H "Content-Type: application/json" \\
  -d '{
    "email": "admin@tenant.test",
    "password": "admin123",
    "tenant_id": "1"
  }'
```

#### 普通用户登录
```bash
curl -X POST "http://localhost:8080/api/v1/auth/test-login" \\
  -H "Content-Type: application/json" \\
  -d '{
    "email": "test@example.com",
    "password": "test123",
    "tenant_id": "1"
  }'
```

### 获取并使用JWT Token

```bash
# 1. 获取Token
JWT_TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/test-login" \\
  -H "Content-Type: application/json" \\
  -d '{
    "email": "admin@system.test",
    "password": "admin123",
    "tenant_id": "0"
  }' | jq -r '.data.access_token')

# 2. 使用Token访问受保护的API
curl -H "Authorization: Bearer $JWT_TOKEN" \\
  "http://localhost:8080/api/v1/users/profile"
```

## 🧪 单元测试集成

在单元测试中使用标准测试用户：

```go
// 获取标准测试用户配置
testUsers := GetStandardTestUsers()
systemAdmin := testUsers[0] // admin@system.test

// 在测试中使用
func TestWithSystemAdmin(t *testing.T) {
    // 使用系统管理员进行测试
    loginReq := dto.TestLoginRequest{
        Email:    "admin@system.test",
        Password: "admin123",
        TenantID: "0",
    }
    
    response, err := userService.TestLogin(ctx, loginReq)
    assert.NoError(t, err)
    assert.NotEmpty(t, response.AccessToken)
}
```

## 📝 权限测试

### 系统租户权限验证
系统租户 (`tenant_id = 0`) 的用户自动拥有所有权限：

```bash
# 系统管理员可以访问任何API，无需检查具体权限
JWT_TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/test-login" \\
  -d '{"email":"admin@system.test","password":"admin123","tenant_id":"0"}' | \\
  jq -r '.data.access_token')

# 访问任何受保护的端点都会成功
curl -H "Authorization: Bearer $JWT_TOKEN" \\
  "http://localhost:8080/api/v1/admin/users"
```

### 租户权限验证
普通租户用户需要具体权限才能访问API：

```bash
# 租户用户需要相应权限才能访问API
JWT_TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/test-login" \\
  -d '{"email":"admin@tenant.test","password":"admin123","tenant_id":"1"}' | \\
  jq -r '.data.access_token')

# 可能因权限不足而返回403
curl -H "Authorization: Bearer $JWT_TOKEN" \\
  "http://localhost:8080/api/v1/admin/users"
```

## 🔄 自动维护

### 定期重置
建议在每次重要开发阶段开始时重置测试用户：

```bash
# 清理并重新创建
go run cmd/migrate/*.go -action=clean-test-users -config=configs/config.dev.yaml
go run cmd/migrate/*.go -action=create-test-users -config=configs/config.dev.yaml
```

### CI/CD集成
在CI/CD流程中自动创建测试用户：

```yaml
# .github/workflows/test.yml
- name: Setup test users
  run: |
    go run cmd/migrate/*.go -action=create-test-users -config=configs/config.test.yaml
```

## ⚠️ 安全注意事项

1. **仅限开发环境**: 测试用户仅应在开发和测试环境中使用
2. **密码安全**: 测试密码较为简单，不适用于生产环境
3. **定期清理**: 在生产部署前确保清理所有测试用户
4. **权限隔离**: 测试不同权限级别时使用不同的测试用户

## 🐛 故障排除

### 登录失败
```bash
# 1. 检查用户是否存在
go run cmd/migrate/*.go -action=list-test-users -config=configs/config.dev.yaml

# 2. 重新创建用户（会更新现有用户密码）
go run cmd/migrate/*.go -action=create-test-users -config=configs/config.dev.yaml

# 3. 检查服务是否运行
curl http://localhost:8080/health
```

### 权限问题
- 系统管理员 (`tenant_id = 0`) 拥有所有权限
- 租户用户需要在数据库中有相应的角色和权限分配
- 使用权限初始化命令确保基础权限存在：
  ```bash
  go run cmd/migrate/*.go -action=migrate -init-permissions -config=configs/config.dev.yaml
  ```

## 📚 相关文档

- [权限系统架构](architecture.md#权限系统)
- [API开发指南](api-guide.md)
- [测试指南](testing-guide.md)
- [数据库迁移](../README.md#数据库迁移)

---

通过使用标准测试用户系统，开发团队可以避免重复的登录调试工作，专注于功能开发和测试。