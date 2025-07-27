# 认证系统 API 接口文档

## 📋 概述

认证API提供用户登录、验证码管理、令牌管理等功能，支持多租户环境下的安全认证。

## 🔧 通用规范

### 1. 请求格式

```http
Content-Type: application/json
Accept: application/json
Authorization: Bearer <access_token>  // 需要认证的接口
X-Tenant-ID: <tenant_uuid>           // 租户标识
```

### 2. 统一响应格式

#### 成功响应
```json
{
  "code": 0,
  "message": "success",
  "data": {
    // 具体数据
  },
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

#### 错误响应
```json
{
  "code": 2005,
  "message": "凭据无效",
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 3. 常用错误码

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

## 🔐 验证码接口

### 1. 获取验证码

**GET** `/api/v1/captcha/generate`

#### 响应示例 (200)

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

#### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| captcha_id | string | 验证码唯一标识 |
| captcha_image | string | Base64编码的验证码图片 |

## 🔑 认证接口

### 1. 用户登录

**POST** `/api/v1/auth/login`

#### 请求参数

```json
{
  "email": "admin@example.com",
  "password": "password123",
  "captcha_id": "bp8RkzOTBEObGLvueygk",
  "captcha_answer": "8849"
}
```

#### 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 用户邮箱 |
| password | string | 是 | 登录密码 |
| captcha_id | string | 是 | 验证码ID |
| captcha_answer | string | 是 | 验证码答案 |

#### 成功响应 (200)

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
      "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
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

#### 错误响应示例

```json
// 验证码错误 (400)
{
  "code": 2011,
  "message": "验证码错误",
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}

// 凭据无效 (401)
{
  "code": 2005,
  "message": "凭据无效",
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}

// 用户被锁定 (403)
{
  "code": 2004,
  "message": "用户被锁定",
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 2. 刷新令牌

**POST** `/api/v1/auth/refresh`

#### 请求参数

```json
{
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g..."
}
```

#### 成功响应 (200)

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

#### 错误响应示例

```json
// 未授权 (401)
{
  "code": 1003,
  "message": "未授权",
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 3. 用户登出

**POST** `/api/v1/auth/logout`

#### 请求头
```http
Authorization: Bearer <access_token>
```

#### 成功响应 (200)

```json
{
  "code": 0,
  "message": "登出成功",
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 4. 获取当前用户信息

**GET** `/api/v1/auth/me`

#### 请求头
```http
Authorization: Bearer <access_token>
```

#### 成功响应 (200)

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

### 5. 获取用户权限信息

**GET** `/api/v1/auth/permissions`

#### 请求头
```http
Authorization: Bearer <access_token>
X-Tenant-ID: <tenant_uuid>
```

#### 成功响应 (200)

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

### 6. 获取可访问租户列表

**GET** `/api/v1/auth/accessible-tenants`

#### 请求头
```http
Authorization: Bearer <access_token>
```

#### 成功响应 (200)

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

## 📝 密码管理

### 1. 修改密码

**PUT** `/api/v1/auth/password`

#### 请求头
```http
Authorization: Bearer <access_token>
```

#### 请求参数

```json
{
  "current_password": "oldPassword123",
  "new_password": "newPassword123",
  "confirm_password": "newPassword123"
}
```

#### 成功响应 (200)

```json
{
  "code": 0,
  "message": "密码修改成功",
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 2. 忘记密码

**POST** `/api/v1/auth/forgot-password`

#### 请求参数

```json
{
  "email": "user@example.com",
  "captcha_id": "bp8RkzOTBEObGLvueygk",
  "captcha_answer": "8849"
}
```

#### 成功响应 (200)

```json
{
  "code": 0,
  "message": "密码重置邮件已发送",
  "data": {
    "email": "user@example.com",
    "expires_in": 1800
  },
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

### 3. 重置密码

**POST** `/api/v1/auth/reset-password`

#### 请求参数

```json
{
  "token": "reset-token-123",
  "new_password": "newPassword123",
  "confirm_password": "newPassword123"
}
```

#### 成功响应 (200)

```json
{
  "code": 0,
  "message": "密码重置成功",
  "trace_id": "1234567890abcdef",
  "timestamp": "2024-01-01T10:00:00Z"
}
```

## 📱 前端集成示例

### 1. 验证码组件使用

```javascript
// 获取验证码
async function getCaptcha() {
  try {
    const response = await axios.get('/api/v1/captcha/generate')
    if (response.data.code === 0) {
      this.captchaId = response.data.data.captcha_id
      this.captchaImage = response.data.data.captcha_image
    }
  } catch (error) {
    console.error('获取验证码失败:', error)
  }
}

// 登录
async function login(loginData) {
  try {
    const response = await axios.post('/api/v1/auth/login', {
      email: loginData.email,
      password: loginData.password,
      captcha_id: this.captchaId,
      captcha_answer: loginData.captchaAnswer
    })
    
    if (response.data.code === 0) {
      // 登录成功
      const { tokens, user, accessible_tenants } = response.data.data
      // 保存token和用户信息
    } else {
      // 处理业务错误
      this.handleError(response.data.code, response.data.message)
    }
  } catch (error) {
    // 网络错误处理
    if (error.response && error.response.data) {
      this.handleError(error.response.data.code, error.response.data.message)
    }
  }
}
```

### 2. 错误处理

```javascript
function handleError(code, message) {
  switch (code) {
    case 2011: // 验证码错误
      this.getCaptcha() // 刷新验证码
      this.$message.error('验证码错误，请重新输入')
      break
    case 2005: // 凭据无效
      this.$message.error('用户名或密码错误')
      break
    case 2004: // 用户被锁定
      this.$message.error('账户已被锁定，请联系管理员')
      break
    default:
      this.$message.error(message || '操作失败')
  }
}
```

## 🔒 安全注意事项

### 1. 验证码安全
- 验证码5分钟过期
- 验证后立即失效
- 验证失败后需要重新获取

### 2. 令牌安全
- Access Token 2小时过期
- Refresh Token 30天过期
- 支持主动令牌失效

### 3. 登录安全
- 5次失败后锁定账户
- 记录登录历史
- IP地址验证

---

## 📋 变更记录

| 版本 | 日期 | 变更内容 |
|------|------|----------|
| 1.0 | 2024-01-01 | 初始版本 |
| 1.1 | 2024-01-01 | 增加验证码接口 |
| 1.2 | 2024-01-01 | 统一响应格式，修正错误码类型 | 