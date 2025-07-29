# 字段权限系统指南

## 概述

字段权限系统是Shield框架的核心安全功能之一，允许管理员精确控制不同角色对数据表字段的访问权限。系统支持三种权限级别：

- **default**: 完全访问权限（可读写）
- **readonly**: 只读权限（只能查看，不能修改）
- **hidden**: 隐藏权限（完全不可见）

## 系统架构

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Permission     │    │  Field           │    │  Response       │
│  Middleware     │───▶│  Permission      │───▶│  Filtering      │
│  (注入权限)      │    │  Service         │    │  (应用权限)      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

### 核心组件

1. **FieldPermission Model**: 字段权限配置表
2. **RoleFieldPermission Model**: 角色字段权限关联表
3. **FieldPermissionService**: 字段权限业务逻辑层
4. **PermissionMiddleware**: 权限注入中间件
5. **FilterResponseByFieldPermissions**: 响应过滤函数

## 快速开始

### 1. 初始化字段权限数据

```bash
# 初始化字段权限配置
go run cmd/migrate/*.go -action=init-field-permissions -config=configs/config.dev.yaml
```

这将为系统中的核心表创建字段权限配置：
- `users` 表：43个字段权限
- `roles` 表：字段权限配置
- `permissions` 表：字段权限配置
- `tenants` 表：字段权限配置

### 2. 配置角色字段权限

```go
// 获取角色的字段权限
roleFieldPermissions, err := fieldPermissionService.GetRoleFieldPermissions(ctx, roleID, tableName)

// 更新角色字段权限
permissions := []models.RoleFieldPermission{
    {
        RoleID:         roleID,
        EntityTable:    "users",
        FieldName:      "password",
        PermissionType: "hidden", // 隐藏密码字段
    },
    {
        RoleID:         roleID,
        EntityTable:    "users", 
        FieldName:      "email",
        PermissionType: "readonly", // 邮箱只读
    },
}
err = fieldPermissionService.UpdateRoleFieldPermissions(ctx, roleID, tableName, permissions)
```

### 3. 在API中应用字段权限

```go
// 在路由中添加字段权限中间件
users.GET("/:uuid", 
    authMiddleware.RequireAuth(),
    authMiddleware.ValidateAPIPermission(), 
    permissionMiddleware.InjectFieldPermissions("users"), // 注入权限
    userHandler.GetUser,
)

// 在Handler中应用过滤
func (h *UserHandler) GetUser(c *gin.Context) {
    // ... 获取用户数据 ...
    user, err := h.userService.GetUserByUUID(c.Request.Context(), uuid)
    if err != nil {
        // ... 错误处理 ...
        return
    }

    // 应用字段权限过滤
    filteredUser := middleware.FilterResponseByFieldPermissions(c, user)
    
    h.responseWriter.Success(c, filteredUser)
}
```

## API 接口

### 获取表字段配置

```http
GET /api/v1/field-permissions/tables/{tableName}/fields
Authorization: Bearer {token}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": "uuid-1",
      "entity_table": "users",
      "field_name": "id",
      "field_label": "用户ID",
      "field_type": "string",
      "default_value": "default",
      "description": "用户唯一标识符",
      "sort_order": 1,
      "is_active": true
    }
  ]
}
```

### 获取角色字段权限

```http
GET /api/v1/field-permissions/roles/{roleId}/{tableName}
Authorization: Bearer {token}
```

### 更新角色字段权限

```http
PUT /api/v1/field-permissions/roles/{roleId}/{tableName}
Authorization: Bearer {token}
Content-Type: application/json

{
  "permissions": [
    {
      "field_name": "password",
      "permission_type": "hidden"
    },
    {
      "field_name": "email", 
      "permission_type": "readonly"
    }
  ]
}
```

## 实际应用示例

### 示例1: 普通用户权限配置

```json
{
  "role": "user",
  "table": "users", 
  "permissions": {
    "id": "default",
    "name": "default", 
    "email": "readonly",
    "phone": "default",
    "password": "hidden",
    "created_at": "readonly",
    "updated_at": "readonly"
  }
}
```

**API响应对比**:

```json
// 系统管理员看到的完整数据
{
  "id": "user-123",
  "name": "张三",
  "email": "zhangsan@example.com",
  "phone": "13800138000",
  "password": "$2a$10$...",  // 管理员可见
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-02T00:00:00Z"
}

// 普通用户看到的过滤数据
{
  "id": "user-123", 
  "name": "张三",
  "email": "zhangsan@example.com", // 只读
  "phone": "13800138000",
  // password 字段被隐藏
  "created_at": "2024-01-01T00:00:00Z", // 只读
  "updated_at": "2024-01-02T00:00:00Z"  // 只读
}
```

### 示例2: 租户管理员权限配置

```json
{
  "role": "tenant_admin",
  "table": "users",
  "permissions": {
    "id": "default",
    "name": "default",
    "email": "default", 
    "phone": "default",
    "password": "hidden",      // 仍然隐藏密码
    "status": "default",       // 可以修改用户状态
    "role_id": "default",      // 可以分配角色
    "created_at": "readonly",
    "updated_at": "readonly"
  }
}
```

### 示例3: 批量数据过滤

```go
// 用户列表API中的字段权限过滤
func (h *UserHandler) ListUsers(c *gin.Context) {
    // ... 获取用户列表 ...
    result, err := h.userService.ListUsers(c.Request.Context(), filter)
    
    // 应用字段权限过滤到整个列表
    filteredUsers := middleware.FilterResponseByFieldPermissions(c, result.Users)
    
    h.responseWriter.Pagination(c, filteredUsers, meta)
}
```

## 权限级别详解

### 1. Default (默认权限)
- **含义**: 完全访问权限
- **行为**: 字段正常显示，支持读写操作
- **适用场景**: 用户基本信息、公开数据

### 2. Readonly (只读权限) 
- **含义**: 只能查看，不能修改
- **行为**: 字段在响应中显示，但前端应禁用编辑
- **适用场景**: 审计字段、系统生成字段

### 3. Hidden (隐藏权限)
- **含义**: 完全不可见
- **行为**: 字段从API响应中完全移除
- **适用场景**: 敏感信息、内部字段

## 性能优化

### 1. 缓存策略

字段权限系统实现了双层缓存：

```go
// Redis缓存（主缓存）
cacheKey := fmt.Sprintf("field_permissions:%s:%s:%s", userID, tenantID, tableName)
cached, err := service.cache.Get(ctx, cacheKey)

// 内存缓存（备用缓存）
if err != nil {
    cached = service.memoryCache.Get(cacheKey)
}
```

### 2. 批量预加载

```go
// 预加载用户的所有表权限
permissions := service.PreloadUserFieldPermissions(ctx, userID, tenantID, []string{"users", "roles", "permissions"})
```

### 3. 权限继承

```go
// 权限继承顺序：用户自定义 > 角色配置 > 系统默认
func (s *service) ResolveFieldPermission(userID, roleID, fieldName string) string {
    // 1. 检查用户自定义权限
    if userPerm := s.getUserFieldPermission(userID, fieldName); userPerm != "" {
        return userPerm
    }
    
    // 2. 检查角色权限配置
    if rolePerm := s.getRoleFieldPermission(roleID, fieldName); rolePerm != "" {
        return rolePerm
    }
    
    // 3. 返回系统默认权限
    return s.getDefaultFieldPermission(fieldName)
}
```

## 测试指南

### 1. 单元测试

```bash
# 运行字段权限服务测试
go test -v ./test/ -run TestFieldPermissionServiceUnitTests

# 测试特定功能
go test -v ./test/ -run "TestFieldPermissionService.*GetTableFields"
go test -v ./test/ -run "TestFieldPermissionService.*UpdateRoleFieldPermissions"
```

### 2. 集成测试

```bash
# 使用标准测试用户测试字段权限
JWT_TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email":"test@example.com",
    "password":"test123", 
    "tenant_id":"1",
    "captcha_id":"dev-bypass",
    "answer":"dev-1234"
  }' | jq -r '.data.access_token')

# 测试用户信息API的字段过滤
curl -H "Authorization: Bearer $JWT_TOKEN" \
  "http://localhost:8080/api/v1/users/user-uuid"
```

### 3. 权限验证测试

```go
func TestFieldPermissionFiltering(t *testing.T) {
    // 设置测试环境
    db, cleanup := SetupTestDB(t)
    defer cleanup()
    
    // 创建测试用户和角色
    user := CreateTestUser(db, 1, "test@example.com")
    role := CreateTestRole(db, 1, "test_role", "测试角色")
    
    // 配置字段权限：隐藏密码字段
    fieldPermissions := []models.RoleFieldPermission{
        {
            RoleID:         role.ID,
            EntityTable:    "users",
            FieldName:      "password", 
            PermissionType: "hidden",
        },
    }
    
    // 验证API响应不包含隐藏字段
    response := callUserAPI(user.UUID)
    assert.NotContains(t, response, "password")
}
```

## 常见问题

### Q1: 如何为新表添加字段权限支持？

**A**: 在 `cmd/migrate/field_permissions.go` 中添加新表配置：

```go
{
    TableName: "new_table",
    Fields: []FieldConfig{
        {
            FieldName:         "id",
            Label:             "ID",
            DefaultPermission: "default",
            Description:       "主键ID",
        },
        // ... 更多字段
    },
},
```

### Q2: 权限不生效怎么办？

**A**: 检查以下几点：
1. 路由是否添加了 `InjectFieldPermissions` 中间件
2. Handler是否调用了 `FilterResponseByFieldPermissions`
3. 缓存是否需要清理
4. 用户角色和权限配置是否正确

### Q3: 如何自定义字段权限逻辑？

**A**: 可以扩展 `filterDataByPermissions` 函数：

```go
func customFilterDataByPermissions(data interface{}, permissions map[string]string, userRole string) interface{} {
    // 自定义权限逻辑
    if userRole == "super_admin" {
        return data // 超级管理员看到所有字段
    }
    
    // 调用默认过滤逻辑
    return filterDataByPermissions(data, permissions)
}
```

## 最佳实践

### 1. 权限设计原则
- **最小权限原则**: 默认给予最小必要权限
- **分层管理**: 系统管理员 > 租户管理员 > 普通用户
- **业务驱动**: 根据实际业务需求配置权限

### 2. 性能优化建议
- 合理使用缓存，避免频繁数据库查询
- 批量处理权限检查，减少网络请求
- 定期清理无效的权限配置

### 3. 安全注意事项
- 敏感字段（如密码）必须设置为 `hidden`
- 审计字段建议设置为 `readonly`
- 定期审查权限配置的合理性

## 总结

字段权限系统为Shield框架提供了细粒度的数据访问控制能力。通过合理配置和使用，可以：

1. **增强安全性**: 防止敏感信息泄露
2. **提升用户体验**: 根据角色显示相关信息
3. **简化前端开发**: 后端统一处理权限逻辑
4. **支持合规要求**: 满足数据保护法规要求

该系统已在用户管理API中完成集成，可作为其他模块的参考实现。