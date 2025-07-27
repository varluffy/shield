# UltraFit 权限系统快速上手指南

## 🚀 5分钟快速开始

本指南帮助开发者快速理解和使用UltraFit权限系统。

## 📋 前置要求

- 已完成项目基础配置
- 数据库已初始化
- 了解基本的HTTP请求

## 🎯 核心概念

UltraFit采用**四层权限控制**：

```
菜单权限 → 按钮权限 → API权限 → 字段权限
   ↓         ↓         ↓         ↓
显示菜单   显示按钮   访问接口   显示字段
```

## ⚡ 快速开始

### 1. 初始化权限数据

```bash
# 运行权限初始化脚本
go run cmd/migrate/main.go -action=permissions
```

### 2. 获取访问令牌

```bash
# 使用测试登录接口
curl -X POST http://localhost:8080/api/v1/auth/test-login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123"
  }'

# 保存返回的access_token
export TOKEN="eyJhbGciOiJIUzI1NiIs..."
```

### 3. 查看权限列表

```bash
# 获取权限树结构
curl -X GET "http://localhost:8080/api/v1/permissions/tree" \
  -H "Authorization: Bearer $TOKEN"
```

### 4. 创建自定义角色

```bash
# 创建新角色
curl -X POST http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "hr_manager",
    "name": "HR经理",
    "description": "人力资源管理角色"
  }'
```

### 5. 分配权限给角色

```bash
# 为角色分配权限（需要先获取permission_ids）
curl -X POST http://localhost:8080/api/v1/roles/2/permissions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "permission_ids": [1, 2, 3, 4, 5]
  }'
```

## 🔧 在代码中使用权限

### 1. 路由权限控制

```go
// 在路由中使用权限中间件
users := api.Group("/users")
users.Use(authMiddleware.RequireAuth())
{
    // 需要特定权限
    users.GET("", authMiddleware.RequirePermission("user_list_api"), handler.ListUsers)
    users.POST("", authMiddleware.RequirePermission("user_create_api"), handler.CreateUser)
    
    // 资源所有者或权限验证
    users.GET("/:uuid", authMiddleware.RequireOwnerOrPermission("uuid", "user_list_api"), handler.GetUser)
}
```

### 2. 在Handler中检查权限

```go
func (h *UserHandler) UpdateUser(c *gin.Context) {
    // 可以在业务逻辑中再次检查权限
    userID := c.GetString("user_id")
    tenantID := c.GetString("tenant_id")
    
    hasPermission, err := h.permissionService.CheckUserPermission(
        ctx, userID, tenantID, "user_update_api")
    
    if !hasPermission {
        h.responseWriter.Error(c, errors.ErrUserPermissionError())
        return
    }
    
    // 执行业务逻辑
}
```

### 3. 字段权限控制

```go
// 注入字段权限
users.Use(permissionMiddleware.InjectFieldPermissions("users"))

// 在Handler中使用字段权限
func (h *UserHandler) GetUser(c *gin.Context) {
    user := getUserFromDB()
    
    // 根据字段权限过滤响应
    response := gin.H{
        "id":   user.ID,
        "name": user.Name,
        "email": user.Email,
    }
    
    // 检查薪资字段权限
    if middleware.HasFieldPermission(c, "salary", "default") {
        response["salary"] = user.Salary
    }
    
    h.responseWriter.Success(c, response)
}
```

## 🎯 常用API接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/permissions` | GET | 获取权限列表 |
| `/api/v1/permissions/tree` | GET | 获取权限树 |
| `/api/v1/roles` | GET/POST | 角色管理 |
| `/api/v1/roles/:id/permissions` | GET/POST | 角色权限管理 |
| `/api/v1/field-permissions/roles/:roleId/:table` | GET/PUT | 字段权限管理 |

## 🔍 权限调试

### 1. 查看用户权限

```bash
# 查看当前用户拥有的权限
curl -X GET "http://localhost:8080/api/v1/auth/permissions" \
  -H "Authorization: Bearer $TOKEN"
```

### 2. 测试权限验证

```bash
# 故意访问没有权限的接口，观察返回的错误信息
curl -X DELETE "http://localhost:8080/api/v1/users/1" \
  -H "Authorization: Bearer $TOKEN"

# 期望返回：403 Forbidden，权限不足
```

### 3. 查看日志

```bash
# 查看权限检查相关日志
grep "Permission check" logs/app.log
```

## ❗ 常见问题

**Q: 用户没有权限访问某个接口？**
```bash
# 1. 检查用户是否有对应角色
# 2. 检查角色是否有对应权限
# 3. 检查权限代码是否正确
```

**Q: 字段权限不生效？**
```bash
# 1. 确认已注入字段权限中间件
# 2. 检查Handler中的字段权限处理逻辑
# 3. 确认角色字段权限配置正确
```

**Q: 权限更改后不生效？**
```bash
# 权限可能被缓存，重启服务或清理缓存
```

## 📚 进阶阅读

- [权限系统完整文档](../business/architecture/permission-system.md)
- [API接口文档](../business/api/permission-api.md)
- [数据库设计文档](../business/database/schema-design.md)

## 🆘 获取帮助

遇到问题时，请按以下顺序查找答案：

1. 查看本快速指南
2. 查看完整权限系统文档
3. 查看代码中的注释和示例
4. 联系开发团队