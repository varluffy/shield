# Shield 服务层测试指南

本指南基于Shield项目的实际服务层单元测试实现，提供服务层测试的最佳实践和具体方法。

## 🎯 概述

Shield项目已实现4个核心服务的完整单元测试覆盖：
- **UserService**: 用户管理服务测试
- **PermissionService**: 权限管理服务测试  
- **RoleService**: 角色管理服务测试
- **BlacklistService**: 黑名单管理服务测试

所有服务层测试都使用真实数据库连接而非Mock，确保测试的真实性和可靠性。

## 🏗️ 服务层测试架构

### 测试文件组织结构

```
test/
├── user_service_test.go       # 用户服务单元测试 (348行)
├── permission_service_test.go # 权限服务单元测试 (321行)
├── role_service_test.go       # 角色服务单元测试 (347行)
├── blacklist_service_test.go  # 黑名单服务单元测试 (391行)
└── test_helpers.go           # 测试辅助方法和标准测试用户
```

### 测试基础模式

每个服务测试文件都遵循相同的基础模式：

```go
func TestServiceUnitTests(t *testing.T) {
    // 1. 设置测试数据库
    db, cleanup := SetupTestDB(t)
    if db == nil {
        return
    }
    defer cleanup()

    // 2. 设置标准测试用户
    testUsers := SetupStandardTestUsers(db)

    // 3. 创建测试组件
    testLogger, err := NewTestLogger()
    require.NoError(t, err)
    components := NewTestComponents(db, testLogger)

    // 4. 运行具体测试用例
    t.Run("Test Feature Success", func(t *testing.T) {
        // 测试逻辑
    })
    
    t.Run("Test Feature Error Cases", func(t *testing.T) {
        // 错误场景测试
    })
}
```

## 🧪 核心测试模式

### 1. 用户服务测试模式

基于 `user_service_test.go` 的实际实现：

#### CRUD操作测试
```go
t.Run("Test CreateUser Success", func(t *testing.T) {
    ctx := context.Background()
    
    // 设置租户上下文（关键！）
    ctx = context.WithValue(ctx, "tenant_id", uint64(1))
    
    req := dto.CreateUserRequest{
        Name:     "测试用户",
        Email:    "newuser@test.com",
        Password: "password123",
        Language: "zh",
        Timezone: "Asia/Shanghai",
    }

    user, err := components.UserService.CreateUser(ctx, req)
    require.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, req.Name, user.Name)
    assert.Equal(t, req.Email, user.Email)
    assert.NotEmpty(t, user.UUID)
})
```

#### 数据验证测试
```go
t.Run("Test CreateUser Invalid Email", func(t *testing.T) {
    ctx := context.Background()
    ctx = context.WithValue(ctx, "tenant_id", uint64(1))
    
    req := dto.CreateUserRequest{
        Name:     "测试用户",
        Email:    "invalid-email", // 无效邮箱格式
        Password: "password123",
    }

    user, err := components.UserService.CreateUser(ctx, req)
    assert.Error(t, err)
    assert.Nil(t, user)
})
```

#### 认证功能测试
```go
t.Run("Test TestLogin Success", func(t *testing.T) {
    ctx := context.Background()
    
    req := dto.TestLoginRequest{
        Email:    "admin@system.test",
        Password: "admin123",
        TenantID: "0",
    }

    response, err := components.UserService.TestLogin(ctx, req)
    require.NoError(t, err)
    assert.NotNil(t, response)
    assert.NotEmpty(t, response.AccessToken)
    assert.NotEmpty(t, response.RefreshToken)
    assert.NotNil(t, response.User)
    assert.Equal(t, req.Email, response.User.Email)
})
```

### 2. 权限服务测试模式

基于 `permission_service_test.go` 的实际实现：

#### 管理员权限测试
```go
t.Run("Test IsSystemAdmin", func(t *testing.T) {
    ctx := context.Background()

    // 测试系统管理员
    systemAdmin := testUsers["admin@system.test"]
    require.NotNil(t, systemAdmin, "系统管理员用户应该存在")

    isSystemAdmin, err := components.PermissionService.IsSystemAdmin(ctx, systemAdmin.UUID)
    require.NoError(t, err)
    assert.True(t, isSystemAdmin, "系统管理员应该返回true")

    // 测试普通用户
    regularUser := testUsers["user@tenant.test"]
    require.NotNil(t, regularUser, "普通用户应该存在")

    isSystemAdmin, err = components.PermissionService.IsSystemAdmin(ctx, regularUser.UUID)
    require.NoError(t, err)
    assert.False(t, isSystemAdmin, "普通用户应该返回false")
})
```

#### 权限查询和过滤测试
```go
t.Run("Test ListPermissions", func(t *testing.T) {
    ctx := context.Background()

    // 测试无过滤条件的权限列表
    filter := make(map[string]interface{})
    permissions, total, err := components.PermissionService.ListPermissions(ctx, filter, 1, 10)
    require.NoError(t, err)
    assert.Greater(t, total, int64(0), "应该有权限数据")
    assert.NotEmpty(t, permissions, "权限列表不应该为空")

    // 测试按类型过滤
    filter["type"] = "api"
    permissions, total, err = components.PermissionService.ListPermissions(ctx, filter, 1, 10)
    require.NoError(t, err)
    
    for _, perm := range permissions {
        assert.Equal(t, "api", perm.Type, "过滤后的权限应该都是API类型")
    }
})
```

### 3. 角色服务测试模式

基于 `role_service_test.go` 的实际实现：

#### 角色权限分配测试
```go
t.Run("Test AssignPermissions Success", func(t *testing.T) {
    ctx := context.Background()

    // 先创建一个角色
    newRole := &models.Role{
        TenantModel: models.TenantModel{TenantID: 1},
        Code:        "permission_role",
        Name:        "权限测试角色",
        Type:        "custom",
        IsActive:    true,
    }

    createdRole, err := components.RoleService.CreateRole(ctx, newRole)
    require.NoError(t, err)

    // 获取一些权限用于分配
    permissions, _, err := components.PermissionService.ListPermissions(ctx, map[string]interface{}{"scope": "tenant"}, 1, 5)
    require.NoError(t, err)
    require.Greater(t, len(permissions), 0, "应该有租户权限可用")

    // 提取权限ID
    permissionIDs := make([]uint64, 0, len(permissions))
    for _, perm := range permissions {
        permissionIDs = append(permissionIDs, perm.ID)
    }

    // 分配权限给角色
    err = components.RoleService.AssignPermissions(ctx, createdRole.ID, permissionIDs)
    require.NoError(t, err)

    // 验证权限已分配
    rolePermissions, err := components.RoleService.GetRolePermissions(ctx, createdRole.ID)
    require.NoError(t, err)
    assert.Greater(t, len(rolePermissions), 0, "角色应该有权限")
})
```

### 4. 黑名单服务测试模式

基于 `blacklist_service_test.go` 的实际实现：

#### 黑名单查询测试
```go
t.Run("Test CheckPhoneMD5 Hit", func(t *testing.T) {
    ctx := context.Background()

    phoneMD5 := generatePhoneMD5("13800138002")
    blacklist := &models.PhoneBlacklist{
        TenantModel: models.TenantModel{TenantID: 1},
        PhoneMD5:    phoneMD5,
        Source:      "manual",
        Reason:      "测试查询命中",
        OperatorID:  1,
        IsActive:    true,
    }

    err := components.BlacklistService.CreateBlacklist(ctx, blacklist)
    require.NoError(t, err)

    // 检查是否在黑名单中
    isBlacklisted, err := components.BlacklistService.CheckPhoneMD5(ctx, 1, phoneMD5)
    require.NoError(t, err)
    assert.True(t, isBlacklisted, "应该在黑名单中")
})
```

#### 服务依赖检查
```go
// 确保黑名单服务可用
if components.BlacklistService == nil {
    t.Skip("黑名单服务不可用，跳过测试")
    return
}
```

## 🎭 测试数据管理

### 标准测试用户使用

所有服务测试都使用标准测试用户系统：

```go
// 获取标准测试用户
testUsers := SetupStandardTestUsers(db)

// 可用的标准测试用户
systemAdmin := testUsers["admin@system.test"]   // 系统管理员 (tenant_id=0)
tenantAdmin := testUsers["admin@tenant.test"]   // 租户管理员 (tenant_id=1)
regularUser := testUsers["user@tenant.test"]    // 普通用户 (tenant_id=1)
testUser := testUsers["test@example.com"]       // 测试用户 (tenant_id=1)
```

### 测试数据清理

每个测试都会自动清理数据：

```go
db, cleanup := SetupTestDB(t)
if db == nil {
    return
}
defer cleanup() // 自动清理测试数据
```

### 测试数据种子

如果需要特定的测试数据，使用专门的种子函数：

```go
// 创建权限测试数据
setupPermissionTestData(db)

// 创建标准测试用户
testUsers := SetupStandardTestUsers(db)
```

## 🚨 错误场景测试

### 输入验证错误

每个服务都应该测试输入验证：

```go
t.Run("Test CreateUser Empty Name", func(t *testing.T) {
    ctx := context.Background()
    ctx = context.WithValue(ctx, "tenant_id", uint64(1))
    
    req := dto.CreateUserRequest{
        Name:     "", // 空名称
        Email:    "test@example.com",
        Password: "password123",
    }

    user, err := components.UserService.CreateUser(ctx, req)
    assert.Error(t, err)
    assert.Nil(t, user)
})
```

### 业务逻辑错误

测试重复数据、权限不足等业务错误：

```go
t.Run("Test CreateUser Duplicate Email", func(t *testing.T) {
    ctx := context.Background()
    ctx = context.WithValue(ctx, "tenant_id", uint64(1))
    
    req := dto.CreateUserRequest{
        Name:     "重复邮箱用户",
        Email:    "user@tenant.test", // 已存在的测试用户邮箱
        Password: "password123",
    }

    user, err := components.UserService.CreateUser(ctx, req)
    assert.Error(t, err)
    assert.Nil(t, user)
    assert.Contains(t, err.Error(), "已存在")
})
```

### 资源不存在错误

测试访问不存在资源的场景：

```go
t.Run("Test GetRoleByID NotFound", func(t *testing.T) {
    ctx := context.Background()

    role, err := components.RoleService.GetRoleByID(ctx, 99999)
    assert.Error(t, err)
    assert.Nil(t, role)
})
```

## 📊 测试覆盖策略

### 全面的功能覆盖

每个服务至少应该覆盖：

1. **CRUD操作**: Create、Read、Update、Delete
2. **查询操作**: 列表查询、条件过滤、分页
3. **业务逻辑**: 特定业务规则验证
4. **权限检查**: 租户隔离、权限验证
5. **错误处理**: 各种异常情况

### 测试用例命名规范

```go
// 格式: TestService_Method_Scenario_ExpectedResult
func TestUserService_CreateUser_ValidInput_Success(t *testing.T) {}
func TestUserService_CreateUser_DuplicateEmail_ReturnsError(t *testing.T) {}
func TestPermissionService_IsSystemAdmin_SystemUser_ReturnsTrue(t *testing.T) {}
```

### 测试分组织

使用 `t.Run` 将相关测试分组：

```go
func TestUserServiceUnitTests(t *testing.T) {
    // 设置代码...
    
    t.Run("Test CreateUser Success", func(t *testing.T) { /* ... */ })
    t.Run("Test CreateUser Duplicate Email", func(t *testing.T) { /* ... */ })
    t.Run("Test CreateUser Missing TenantID", func(t *testing.T) { /* ... */ })
}

func TestUserServiceValidation(t *testing.T) {
    // 专门的验证测试分组
    t.Run("Test CreateUser Invalid Email", func(t *testing.T) { /* ... */ })
    t.Run("Test CreateUser Empty Name", func(t *testing.T) { /* ... */ })
    t.Run("Test CreateUser Short Password", func(t *testing.T) { /* ... */ })
}
```

## 🔧 测试运行和调试

### 运行特定服务测试

```bash
# 运行所有服务层测试
go test -v ./test/ -run ".*ServiceUnitTests"

# 运行特定服务测试
go test -v ./test/ -run TestUserServiceUnitTests
go test -v ./test/ -run TestPermissionServiceUnitTests
go test -v ./test/ -run TestRoleServiceUnitTests
go test -v ./test/ -run TestBlacklistServiceUnitTests

# 运行特定测试场景
go test -v ./test/ -run "TestUserService.*CreateUser"
go test -v ./test/ -run "TestPermissionService.*IsSystemAdmin"
```

### 测试调试技巧

```go
// 添加调试输出
t.Logf("用户创建结果: %+v", user)
t.Logf("错误信息: %v", err)

// 使用 require 进行必要条件检查
require.NotNil(t, testUsers["admin@system.test"], "系统管理员用户应该存在")

// 验证具体的错误信息
assert.Contains(t, err.Error(), "已存在", "错误信息应该明确说明邮箱已存在")
```

## ✅ 测试清单

### 新服务测试开发清单

当添加新服务时，确保包含以下测试：

#### 基础功能测试
- [ ] 服务创建和初始化测试
- [ ] CRUD操作的成功场景测试
- [ ] 查询和列表功能测试
- [ ] 分页功能测试（如适用）

#### 数据验证测试
- [ ] 必填字段验证测试
- [ ] 数据格式验证测试（邮箱、电话等）
- [ ] 数据长度限制测试
- [ ] 特殊字符处理测试

#### 业务逻辑测试
- [ ] 唯一性约束测试（如邮箱、代码等）
- [ ] 权限检查测试
- [ ] 租户隔离测试
- [ ] 状态转换测试（如激活/禁用）

#### 错误场景测试
- [ ] 输入参数错误测试
- [ ] 资源不存在测试
- [ ] 依赖服务不可用测试
- [ ] 数据库连接失败测试

#### 集成测试
- [ ] 与其他服务的集成测试
- [ ] 事务处理测试
- [ ] 缓存功能测试（如适用）
- [ ] 外部API调用测试（如适用）

## 📚 相关文档

- 📋 [测试速查手册](./testing-cheatsheet.md) - 快速参考常用测试命令和模式 ⚡
- 🧪 [主要测试指南](./testing-guide.md) - 完整的测试策略和工具使用
- 👥 [测试用户管理](./test-users.md) - 标准测试用户系统使用指南
- 📊 [测试系统重构报告](../reports/test-system-refactoring.md) - 测试系统演进历程
- 🏗️ [架构设计指南](./architecture.md) - 可测试的架构设计原则

---

**最佳实践提醒**: Shield项目的服务层测试注重真实性和全面性，使用真实数据库连接而非Mock，确保测试能够发现实际的集成问题。在编写新的服务测试时，请参考现有的4个服务测试文件的实现模式，保持测试风格的一致性。