# Shield 测试速查手册

快速参考指南，包含最常用的测试命令和模式。

## 🚀 快速命令

### 运行测试
```bash
# 所有测试
make test

# 完整验证周期 (必须！修改代码后)
make wire && make test

# 服务层单元测试
go test -v ./test/ -run ".*ServiceUnitTests"

# 特定服务测试
go test -v ./test/ -run TestUserServiceUnitTests
go test -v ./test/ -run TestPermissionServiceUnitTests
go test -v ./test/ -run TestRoleServiceUnitTests
go test -v ./test/ -run TestFieldPermissionServiceUnitTests
go test -v ./test/ -run TestBlacklistServiceUnitTests

# 集成测试
go test -v ./test/ -run TestPermissionSystemIntegration
go test -v ./test/ -run TestPermissionFilteringUnit

# 覆盖率报告
go test -v -cover ./test/
go test -v -coverprofile=coverage.out ./test/
go tool cover -html=coverage.out
```

### 调试和故障排除
```bash
# 详细输出
go test -v ./test/ -run TestFailingTest

# 竞态条件检测
go test -race ./test/

# 特定功能测试
go test -v ./test/ -run TestCaptcha
go test -v ./test/ -run TestPermission
go test -v ./test/ -run TestFieldPermission

# 检查服务状态
make status
make stop-service
```

## 👥 标准测试用户

### 可用用户
```bash
# 系统管理员 (推荐用于开发)
Email: admin@system.test
Password: admin123  
Tenant: 0 (系统租户，绕过权限检查)

# 租户管理员
Email: admin@tenant.test
Password: admin123
Tenant: 1

# 普通用户
Email: user@tenant.test  
Password: user123
Tenant: 1

# 测试用户
Email: test@example.com
Password: test123
Tenant: 1
```

### 快速认证
```bash
# 获取系统管理员Token (开发推荐)
JWT_TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email":"admin@system.test",
    "password":"admin123",
    "tenant_id":"0",
    "captcha_id":"dev-bypass",
    "answer":"dev-1234"
  }' | jq -r '.data.access_token')

# 测试API访问
curl -H "Authorization: Bearer $JWT_TOKEN" "http://localhost:8080/api/v1/users/profile"

# 租户用户Token
JWT_TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email":"test@example.com",
    "password":"test123",
    "tenant_id":"1",
    "captcha_id":"dev-bypass",
    "answer":"dev-1234"
  }' | jq -r '.data.access_token')
```

## 📝 测试代码模板

### 新服务测试模板
```go
func TestNewServiceUnitTests(t *testing.T) {
    // 1. 标准设置
    db, cleanup := SetupTestDB(t)
    if db == nil { return }
    defer cleanup()
    
    testUsers := SetupStandardTestUsers(db)
    testLogger, err := NewTestLogger()
    require.NoError(t, err)
    components := NewTestComponents(db, testLogger)
    
    // 2. 成功场景测试
    t.Run("Test Create Success", func(t *testing.T) {
        // 测试成功创建
    })
    
    // 3. 错误场景测试
    t.Run("Test Create Invalid Input", func(t *testing.T) {
        // 测试输入验证错误
    })
    
    // 4. 业务逻辑测试
    t.Run("Test Business Logic", func(t *testing.T) {
        // 测试业务规则
    })
}
```

### 认证测试模板
```go
// 获取测试用户
systemAdmin := testUsers["admin@system.test"]
tenantUser := testUsers["user@tenant.test"]

// 生成JWT Token
token, err := GenerateTestJWT(components, systemAdmin.UUID, "0")
require.NoError(t, err)

// 创建认证头
authHeaders := CreateAuthHeader(token)
```

### 租户隔离测试模板
```go
t.Run("Test Tenant Isolation", func(t *testing.T) {
    ctx := context.Background()
    ctx = context.WithValue(ctx, "tenant_id", uint64(1))
    
    // 创建租户1的数据
    data1 := &models.Entity{
        TenantModel: models.TenantModel{TenantID: 1},
        Name: "租户1数据",
    }
    err := service.Create(ctx, data1)
    require.NoError(t, err)
    
    // 切换到租户2上下文
    ctx2 := context.WithValue(context.Background(), "tenant_id", uint64(2))
    
    // 验证租户2看不到租户1的数据
    result, err := service.GetByTenant(ctx2, 2)
    require.NoError(t, err)
    assert.Empty(t, result, "租户2不应该看到租户1的数据")
})
```

### 字段权限测试模板
```go
t.Run("Test Field Permission Filtering", func(t *testing.T) {
    // 设置角色字段权限：隐藏密码字段
    permissions := []models.RoleFieldPermission{
        {
            RoleID:         roleID,
            EntityTable:    "users",
            FieldName:      "password",
            PermissionType: "hidden",
        },
    }
    
    err := fieldPermissionService.UpdateRoleFieldPermissions(ctx, roleID, "users", permissions)
    require.NoError(t, err)
    
    // 验证API响应不包含隐藏字段
    response := callUserAPI(user.UUID)
    assert.NotContains(t, response, "password", "密码字段应该被隐藏")
})
```

## 🏗️ 测试设置助手

### 数据库设置
```go
// 基本设置
db, cleanup := SetupTestDB(t)
if db == nil {
    return // 数据库不可用时跳过
}
defer cleanup()

// 标准测试用户
testUsers := SetupStandardTestUsers(db)

// 测试组件
testLogger, err := NewTestLogger()
require.NoError(t, err)
components := NewTestComponents(db, testLogger)
```

### 常用断言
```go
// 基本断言
assert.NoError(t, err)
assert.Error(t, err)
assert.Nil(t, result)
assert.NotNil(t, result)
assert.Empty(t, list)
assert.NotEmpty(t, list)

// 比较断言
assert.Equal(t, expected, actual)
assert.NotEqual(t, expected, actual)
assert.Greater(t, actual, expected)
assert.Contains(t, str, substring)

// 必要条件 (失败时终止测试)
require.NoError(t, err)
require.NotNil(t, user, "用户不应该为nil")
```

## 🐛 常见问题解决

### 数据库连接问题
```bash
# 检查MySQL服务
systemctl status mysql

# 通过MCP工具测试连接 (推荐)
SELECT 1;

# ❌ 不要使用这些命令:
# mysql -u root -p shield
# mysql shield -e "SELECT 1;"
```

### Wire依赖注入问题
```bash
# 症状: 构建失败，提示缺少providers
# 解决: 重新生成依赖注入
make wire

# 检查: 确认新的构造函数已添加到对应的ProviderSet
```

### 端口占用问题
```bash
# 症状: "port already in use"
# 解决:
make stop-service      # 清理停止
make kill-port         # 强制杀死端口8080进程
make status           # 验证状态
```

### 测试数据干扰
```bash
# 症状: 测试相互影响
# 解决: 确保每个测试都有独立的cleanup
defer cleanup()

# 或者使用事务回滚
db.Transaction(func(tx *gorm.DB) error {
    // 测试逻辑
    return errors.New("rollback") // 强制回滚
})
```

## 📊 测试覆盖标准

### 必测场景
- ✅ **成功场景**: 正常输入的成功执行
- ✅ **输入验证**: 空值、无效格式、超长字符串
- ✅ **业务逻辑**: 重复创建、权限检查、状态转换
- ✅ **错误处理**: 资源不存在、依赖失败
- ✅ **租户隔离**: 多租户数据访问控制

### 覆盖率要求
| 层级 | 最低要求 | 目标 |
|------|---------|------|
| 核心服务层 | 90% | 95% |
| Handler层 | 60% | 80% |
| Repository层 | 70% | 85% |
| 权限系统 | 95% | 100% |

## 🎯 最佳实践提醒

### DO ✅
- 使用标准测试用户避免重复创建
- 使用真实数据库连接确保集成可靠性
- 每个测试都包含cleanup逻辑
- 使用描述性的测试名称
- 测试租户隔离和权限控制
- 修改代码后运行 `make wire && make test`

### DON'T ❌
- 不要使用直接MySQL命令 (仅使用MCP工具)
- 不要在测试间共享数据
- 不要过度使用Mock (优先真实连接)
- 不要忽略错误场景测试
- 不要跳过租户上下文设置
- 不要忘记运行Wire重新生成依赖

## 📚 相关文档
- 📖 [完整测试指南](./testing-guide.md)
- 🔧 [服务层测试指南](./service-testing-guide.md)  
- 👥 [测试用户指南](./test-users.md)
- 🏗️ [架构设计指南](./architecture.md)

---
**💡 提示**: 这是快速参考，详细信息请参考对应的完整指南文档。