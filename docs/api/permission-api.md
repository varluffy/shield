# UltraFit 权限系统 API 接口文档

## 1. 通用规范

### 1.1 请求格式
- **Base URL**: `/api/v1`
- **Content-Type**: `application/json`
- **认证方式**: `Authorization: Bearer <access_token>`
- **租户标识**: `X-Tenant-ID: <tenant_uuid>`

### 1.2 响应格式
```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 1.3 错误响应
```json
{
  "code": 1002,
  "message": "参数验证失败",
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 1.4 常用错误码

| 错误码 | 说明 | HTTP状态码 |
|--------|------|------------|
| 0 | 成功 | 200 |
| 1001 | 无效请求 | 400 |
| 1002 | 参数验证失败 | 400 |
| 1003 | 未授权 | 401 |
| 1004 | 禁止访问 | 403 |
| 2001 | 用户不存在 | 404 |
| 2004 | 用户被锁定 | 403 |
| 2005 | 凭据无效 | 401 |
| 2010 | 需要验证码 | 400 |
| 2011 | 验证码错误 | 400 |
| 2012 | 验证码已过期 | 400 |

## 2. 认证相关接口

### 2.1 获取验证码
**GET** `/api/v1/captcha/generate`

**响应数据**:
```json
{
  "code": 0,
  "message": "验证码生成成功",
  "data": {
    "captcha_id": "bp8RkzOTBEObGLvueygk",
    "captcha_image": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAJYAAAA..."
  },
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 2.2 用户登录
**POST** `/api/v1/auth/login`

**请求参数**:
```json
{
  "email": "admin@example.com",
  "password": "password123",
  "captcha_id": "bp8RkzOTBEObGLvueygk",
  "captcha_answer": "8849"
}
```

**响应数据**:
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
      "is_system_admin": true,
      "last_login_at": "2024-01-01T10:00:00Z"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIs...",
      "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g...",
      "token_type": "Bearer",
      "expires_in": 7200
    },
    "accessible_tenants": [
      {
        "id": 0,
        "uuid": "system",
        "name": "系统管理",
        "code": "system"
      },
      {
        "id": 1,
        "uuid": "550e8400-e29b-41d4-a716-446655440000",
        "name": "租户A",
        "code": "tenant_a",
        "status": "active"
      }
    ]
  },
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 2.3 刷新令牌
**POST** `/api/v1/auth/refresh`

**请求参数**:
```json
{
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g..."
}
```

**响应数据**:
```json
{
  "code": 0,
  "message": "刷新成功",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 7200
  },
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 2.4 用户登出
**POST** `/api/v1/auth/logout`

**响应数据**:
```json
{
  "code": 0,
  "message": "登出成功",
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 2.5 获取当前用户信息
**GET** `/api/v1/auth/me`

**响应数据**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "tenant_id": 0,
    "is_system_admin": true,
    "last_login_at": "2024-01-01T10:00:00Z"
  },
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 2.6 获取用户权限信息
**GET** `/api/v1/auth/permissions`

**响应数据**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": {
    "menus": [
      {
        "code": "system_management",
        "name": "系统管理",
        "path": "/system",
        "children": [
          {
            "code": "user_management",
            "name": "用户管理",
            "path": "/system/users",
            "buttons": ["user_create_btn", "user_edit_btn"]
          }
        ]
      }
    ],
    "buttons": ["user_create_btn", "user_edit_btn", "user_delete_btn"],
    "apis": ["user_create_api", "user_list_api", "user_update_api"]
  },
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 2.7 获取可访问租户列表
**GET** `/api/v1/auth/accessible-tenants`

**响应数据**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": [
    {
      "id": 0,
      "uuid": "system",
      "name": "系统管理",
      "code": "system"
    },
    {
      "id": 1,
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "name": "租户A",
      "code": "tenant_a",
      "status": "active"
    }
  ],
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

## 3. 系统管理接口

### 3.1 租户管理

#### 3.1.1 创建租户
**POST** `/api/v1/system/tenants`

**请求参数**:
```json
{
  "tenant_code": "tenant_b",
  "tenant_name": "租户B",
  "max_users": 50,
  "admin_email": "admin@tenant-b.com",
  "admin_password": "AdminPassword123"
}
```

**响应数据**:
```json
{
  "code": 0,
  "message": "租户创建成功",
  "data": {
    "id": 2,
    "uuid": "550e8400-e29b-41d4-a716-446655440001",
    "tenant_code": "tenant_b",
    "tenant_name": "租户B",
    "status": "active",
    "max_users": 50,
    "admin_user": {
      "id": 10,
      "username": "admin",
      "email": "admin@tenant-b.com"
    }
  },
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

#### 3.1.2 获取租户列表
**GET** `/api/v1/system/tenants`

**查询参数**:
- `page`: 页码（默认：1）
- `limit`: 每页数量（默认：20）
- `status`: 状态筛选

**响应数据**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "tenant_code": "tenant_a",
      "tenant_name": "租户A",
      "status": "active",
      "max_users": 100,
      "current_users": 25,
      "created_at": "2024-01-01T10:00:00Z"
    }
  ],
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

#### 3.1.3 更新租户信息
**PUT** `/api/v1/system/tenants/{tenant_id}`

**请求参数**:
```json
{
  "tenant_name": "租户A（更新）",
  "max_users": 150,
  "status": "active"
}
```

**响应数据**:
```json
{
  "code": 0,
  "message": "更新成功",
  "data": {
    "id": 1,
    "tenant_name": "租户A（更新）",
    "max_users": 150,
    "status": "active",
    "updated_at": "2024-01-01T10:30:00Z"
  },
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:30:00Z"
}
```

#### 3.1.4 删除租户
**DELETE** `/api/v1/system/tenants/{tenant_id}`

**响应数据**:
```json
{
  "code": 0,
  "message": "租户删除成功",
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

## 4. 租户管理接口

### 4.1 用户管理

#### 4.1.1 创建用户
**POST** `/api/v1/users`

**请求参数**:
```json
{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "Password123",
  "role_ids": [2]
}
```

**响应数据**:
```json
{
  "code": 0,
  "message": "用户创建成功",
  "data": {
    "id": 11,
    "username": "newuser",
    "email": "newuser@example.com",
    "tenant_id": 1,
    "status": "active",
    "roles": [
      {
        "id": 2,
        "role_code": "tenant_admin",
        "role_name": "租户管理员"
      }
    ]
  },
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

#### 4.1.2 获取用户列表
**GET** `/api/v1/users`

**查询参数**:
- `page`: 页码（默认：1）
- `limit`: 每页数量（默认：20）
- `status`: 状态筛选
- `role_id`: 角色筛选

**响应数据**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": [
    {
      "id": 11,
      "username": "newuser",
      "email": "newuser@example.com",
      "status": "active",
      "roles": [
        {
          "id": 2,
          "role_code": "tenant_admin",
          "role_name": "租户管理员"
        }
      ],
      "last_login_at": "2024-01-01T09:00:00Z",
      "created_at": "2024-01-01T08:00:00Z"
    }
  ],
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

#### 4.1.3 更新用户信息
**PUT** `/api/v1/users/{user_id}`

**请求参数**:
```json
{
  "username": "updateduser",
  "email": "updated@example.com",
  "status": "active"
}
```

**响应数据**:
```json
{
  "code": 0,
  "message": "用户更新成功",
  "data": {
    "id": 11,
    "username": "updateduser",
    "email": "updated@example.com",
    "status": "active",
    "updated_at": "2024-01-01T10:30:00Z"
  },
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:30:00Z"
}
```

#### 4.1.4 删除用户
**DELETE** `/api/v1/users/{user_id}`

**响应数据**:
```json
{
  "code": 0,
  "message": "用户删除成功",
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 4.2 角色管理

#### 4.2.1 获取角色列表
**GET** `/api/v1/roles`

**响应数据**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "role_code": "system_admin",
      "role_name": "系统管理员",
      "level": 1,
      "is_system_role": true,
      "permissions_count": 50
    },
    {
      "id": 2,
      "role_code": "tenant_admin",
      "role_name": "租户管理员",
      "level": 10,
      "is_system_role": true,
      "permissions_count": 25
    }
  ],
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

#### 4.2.2 创建自定义角色
**POST** `/api/v1/roles`

**请求参数**:
```json
{
  "role_code": "custom_manager",
  "role_name": "自定义管理员",
  "level": 20,
  "description": "自定义角色描述"
}
```

**响应数据**:
```json
{
  "code": 0,
  "message": "角色创建成功",
  "data": {
    "id": 3,
    "role_code": "custom_manager",
    "role_name": "自定义管理员",
    "level": 20,
    "is_system_role": false,
    "description": "自定义角色描述"
  },
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 4.3 权限管理

#### 4.3.1 获取权限列表
**GET** `/api/v1/permissions`

**响应数据**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": [
    {
      "id": 1,
      "permission_code": "user_management",
      "permission_name": "用户管理",
      "permission_type": "menu",
      "parent_code": null,
      "resource_path": "/users",
      "children": [
        {
          "id": 2,
          "permission_code": "user_create_btn",
          "permission_name": "创建用户按钮",
          "permission_type": "button",
          "parent_code": "user_management"
        },
        {
          "id": 3,
          "permission_code": "user_create_api",
          "permission_name": "创建用户接口",
          "permission_type": "api",
          "parent_code": "user_management",
          "resource_path": "/api/v1/users",
          "method": "POST"
        }
      ]
    }
  ],
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

#### 4.3.2 为角色分配权限
**POST** `/api/v1/roles/{role_id}/permissions`

**请求参数**:
```json
{
  "permission_ids": [1, 2, 3, 4, 5]
}
```

**响应数据**:
```json
{
  "code": 0,
  "message": "权限分配成功",
  "data": {
    "role_id": 3,
    "assigned_permissions": 5,
    "updated_at": "2024-01-01T10:00:00Z"
  },
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

#### 4.3.3 获取角色权限
**GET** `/api/v1/roles/{role_id}/permissions`

**响应数据**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": {
    "role": {
      "id": 3,
      "role_code": "custom_manager",
      "role_name": "自定义管理员"
    },
    "permissions": [
      {
        "id": 1,
        "permission_code": "user_management",
        "permission_name": "用户管理",
        "permission_type": "menu"
      },
      {
        "id": 2,
        "permission_code": "user_create_btn",
        "permission_name": "创建用户按钮",
        "permission_type": "button"
      }
    ]
  },
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

## 5. 错误处理示例

### 5.1 参数验证错误
```json
{
  "code": 1002,
  "message": "参数验证失败",
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 5.2 权限不足错误
```json
{
  "code": 1004,
  "message": "禁止访问",
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 5.3 资源不存在错误
```json
{
  "code": 2001,
  "message": "用户不存在",
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

## 6. 前端集成示例

### 6.1 错误处理
```javascript
function handleApiResponse(response) {
  if (response.data.code === 0) {
    // 成功处理
    return response.data.data
  } else {
    // 错误处理
    switch (response.data.code) {
      case 1003:
        // 未授权，跳转登录
        router.push('/login')
        break
      case 1004:
        // 权限不足
        this.$message.error('权限不足')
        break
      case 2011:
        // 验证码错误，刷新验证码
        this.refreshCaptcha()
        break
      default:
        this.$message.error(response.data.message)
    }
    throw new Error(response.data.message)
  }
}
```

### 6.2 权限控制
```javascript
// 检查按钮权限
function hasButtonPermission(buttonCode) {
  const userPermissions = store.getters.userPermissions
  return userPermissions.buttons.includes(buttonCode)
}

// 检查API权限
function hasApiPermission(path, method) {
  const userPermissions = store.getters.userPermissions
  const apiCode = `${path}_${method.toLowerCase()}_api`
  return userPermissions.apis.includes(apiCode)
}
```

---

## 📋 总结

本API文档涵盖了UltraFit权限系统的所有核心接口，包括：

1. **认证管理**：登录、验证码、令牌管理
2. **系统管理**：租户管理（系统管理员）
3. **租户管理**：用户、角色、权限管理（租户管理员）
4. **统一格式**：所有接口使用统一的响应格式
5. **错误处理**：详细的错误码和处理示例

该设计确保了多租户环境下的数据隔离和权限控制，为前端开发提供了清晰的接口规范。 