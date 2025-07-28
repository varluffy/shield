# UltraFit API 开发指南

本指南详细说明了 UltraFit 项目的 API 设计规范、开发标准和最佳实践，确保 API 的一致性、安全性和可维护性。

## 🎯 API 设计原则

### 核心原则
- **RESTful 设计**: 遵循 REST 架构风格
- **统一响应格式**: 标准化的请求/响应结构
- **安全第一**: 完整的认证授权机制
- **多租户支持**: 租户隔离和上下文传递
- **可观测性**: 完整的链路追踪和日志记录

### 设计标准
```
📋 统一性    - 命名、格式、错误处理统一
🔒 安全性    - 认证、授权、输入验证、HTTPS
📊 可观测性  - 日志、指标、追踪、监控
🏗️ 可扩展性  - 版本控制、向后兼容、模块化
⚡ 性能     - 缓存、分页、异步处理
```

## 🌐 API 规范标准

### URL 设计规范

**基础路径结构**:
```
https://api.example.com/api/v1/{resource}
                       │   │  └── 资源名称（复数）
                       │   └── 版本号
                       └── API 前缀
```

**标准 URL 模式**:
```bash
# 资源集合操作
GET    /api/v1/users              # 获取用户列表
POST   /api/v1/users              # 创建新用户
GET    /api/v1/users/search       # 搜索用户

# 单个资源操作
GET    /api/v1/users/{id}         # 获取单个用户
PUT    /api/v1/users/{id}         # 更新用户（完整更新）
PATCH  /api/v1/users/{id}         # 部分更新用户
DELETE /api/v1/users/{id}         # 删除用户

# 嵌套资源操作
GET    /api/v1/users/{id}/roles   # 获取用户角色
POST   /api/v1/users/{id}/roles   # 为用户分配角色
```

**命名约定**:
- 使用小写字母和连字符
- 资源名使用复数形式
- 避免动词，使用 HTTP 方法表示操作
- 嵌套深度不超过 3 层

### HTTP 方法规范

| 方法 | 用途 | 幂等性 | 安全性 | 响应码 |
|------|------|--------|--------|--------|
| GET | 查询资源 | ✅ | ✅ | 200, 404 |
| POST | 创建资源 | ❌ | ❌ | 201, 400 |
| PUT | 完整更新 | ✅ | ❌ | 200, 404 |
| PATCH | 部分更新 | ❌ | ❌ | 200, 404 |
| DELETE | 删除资源 | ✅ | ❌ | 204, 404 |

### 请求规范

**请求头标准**:
```http
Content-Type: application/json
Accept: application/json
Authorization: Bearer <access_token>
X-Tenant-ID: <tenant_uuid>
X-Request-ID: <request_id>
```

**请求体示例**:
```json
{
  "name": "张三",
  "email": "zhangsan@example.com",
  "phone": "+86-13800138000",
  "metadata": {
    "source": "web",
    "campaign": "spring_promotion"
  }
}
```

**查询参数规范**:
```bash
# 分页参数
?page=1&page_size=20

# 排序参数
?sort=created_at&order=desc

# 过滤参数
?status=active&role=admin&created_after=2024-01-01

# 字段选择
?fields=id,name,email

# 搜索参数
?q=keyword&search_in=name,email
```

## 📄 响应格式标准

### 统一响应结构

```json
{
  "code": 0,
  "message": "success",
  "data": {
    // 具体业务数据
  },
  "meta": {
    "trace_id": "1234567890abcdef",
    "timestamp": "2024-01-01T10:00:00Z",
    "version": "v1.0.0"
  }
}
```

### 成功响应示例

**单个资源**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": {
    "id": "user_123",
    "name": "张三",
    "email": "zhangsan@example.com",
    "created_at": "2024-01-01T10:00:00Z"
  },
  "meta": {
    "trace_id": "abc123",
    "timestamp": "2024-01-01T10:00:00Z"
  }
}
```

**资源列表**:
```json
{
  "code": 0,
  "message": "获取成功",
  "data": {
    "items": [
      {
        "id": "user_123",
        "name": "张三",
        "email": "zhangsan@example.com"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5
    }
  },
  "meta": {
    "trace_id": "abc123",
    "timestamp": "2024-01-01T10:00:00Z"
  }
}
```

### 错误响应标准

**错误响应结构**:
```json
{
  "code": 1002,
  "message": "参数验证失败",
  "errors": [
    {
      "field": "email",
      "message": "邮箱格式不正确",
      "code": "invalid_format"
    },
    {
      "field": "phone",
      "message": "手机号不能为空",
      "code": "required"
    }
  ],
  "meta": {
    "trace_id": "abc123",
    "timestamp": "2024-01-01T10:00:00Z"
  }
}
```

**标准错误码**:
```go
// 成功
const (
    SUCCESS = 0
)

// 系统错误 (1000-1999)
const (
    INVALID_REQUEST     = 1001  // 无效请求
    VALIDATION_FAILED   = 1002  // 参数验证失败
    UNAUTHORIZED        = 1003  // 未授权
    FORBIDDEN          = 1004  // 禁止访问
    NOT_FOUND          = 1005  // 资源不存在
    INTERNAL_ERROR     = 1006  // 内部服务器错误
)

// 业务错误 (2000-2999)
const (
    USER_NOT_FOUND     = 2001  // 用户不存在
    EMAIL_EXISTS       = 2002  // 邮箱已存在
    INVALID_PASSWORD   = 2003  // 密码错误
    USER_LOCKED        = 2004  // 用户被锁定
    INVALID_CREDENTIALS = 2005  // 凭据无效
)

// 验证码错误 (2010-2019)
const (
    CAPTCHA_REQUIRED   = 2010  // 需要验证码
    CAPTCHA_INVALID    = 2011  // 验证码错误
    CAPTCHA_EXPIRED    = 2012  // 验证码已过期
)
```

## 🔒 认证与授权

### JWT 认证机制

**Token 结构**:
```json
{
  "header": {
    "alg": "RS256",
    "typ": "JWT"
  },
  "payload": {
    "user_id": "user_123",
    "tenant_id": "tenant_456",
    "roles": ["admin", "user"],
    "permissions": ["user:read", "user:write"],
    "exp": 1704067200,
    "iat": 1704063600,
    "iss": "shield"
  }
}
```

**认证流程**:
```go
// 1. 认证中间件
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractBearerToken(c)
        if token == "" {
            response.Unauthorized(c, "缺少认证令牌")
            c.Abort()
            return
        }
        
        claims, err := validateJWT(token)
        if err != nil {
            response.Unauthorized(c, "无效令牌")
            c.Abort()
            return
        }
        
        // 设置用户上下文
        c.Set("user_id", claims.UserID)
        c.Set("tenant_id", claims.TenantID)
        c.Set("permissions", claims.Permissions)
        c.Next()
    }
}

// 2. 权限验证
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        permissions, exists := c.Get("permissions")
        if !exists || !hasPermission(permissions, permission) {
            response.Forbidden(c, "权限不足")
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 多租户支持

**租户上下文传递**:
```go
// 中间件设置租户上下文
func TenantMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID := c.GetHeader("X-Tenant-ID")
        if tenantID == "" {
            // 从 JWT 中获取租户信息
            tenantID = getUserTenantFromJWT(c)
        }
        
        c.Set("tenant_id", tenantID)
        c.Next()
    }
}

// 服务层使用租户上下文
func (s *userService) GetUsers(ctx context.Context) ([]*models.User, error) {
    tenantID := getTenantIDFromContext(ctx)
    return s.userRepo.GetByTenant(ctx, tenantID)
}
```

## 🛠️ API 开发实践

### Handler 层实现

**标准 Handler 结构**:
```go
type UserHandler struct {
    userService services.UserService
    logger      logger.Logger
}

// 获取用户列表
func (h *UserHandler) GetUsers(c *gin.Context) {
    // 1. 参数绑定和验证
    var req GetUsersRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        response.ValidationError(c, err)
        return
    }
    
    // 2. 调用服务层
    users, total, err := h.userService.GetUsers(
        c.Request.Context(),
        req.ToServiceParams(),
    )
    if err != nil {
        h.logger.ErrorWithTrace(c.Request.Context(), "获取用户列表失败", zap.Error(err))
        response.InternalError(c, "获取用户列表失败")
        return
    }
    
    // 3. 构建响应
    resp := GetUsersResponse{
        Items: make([]UserInfo, len(users)),
        Pagination: PaginationInfo{
            Page:       req.Page,
            PageSize:   req.PageSize,
            Total:      total,
            TotalPages: (total + req.PageSize - 1) / req.PageSize,
        },
    }
    
    for i, user := range users {
        resp.Items[i] = UserInfo{
            ID:        user.ID,
            Name:      user.Name,
            Email:     user.Email,
            CreatedAt: user.CreatedAt,
        }
    }
    
    response.Success(c, "获取成功", resp)
}
```

### DTO 设计模式

**请求 DTO**:
```go
// 请求参数结构
type CreateUserRequest struct {
    Name     string            `json:"name" binding:"required,min=2,max=50"`
    Email    string            `json:"email" binding:"required,email"`
    Phone    string            `json:"phone" binding:"required,phone"`
    Metadata map[string]string `json:"metadata"`
}

// 验证规则
func (r *CreateUserRequest) Validate() error {
    if !isValidPhone(r.Phone) {
        return errors.New("手机号格式不正确")
    }
    return nil
}

// 转换为领域模型
func (r *CreateUserRequest) ToModel(tenantID string) *models.User {
    return &models.User{
        TenantID: tenantID,
        Name:     r.Name,
        Email:    r.Email,
        Phone:    r.Phone,
        Metadata: r.Metadata,
    }
}
```

**响应 DTO**:
```go
// 响应数据结构
type UserInfo struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}

// 从领域模型构建
func NewUserInfo(user *models.User) UserInfo {
    return UserInfo{
        ID:        user.ID,
        Name:      user.Name,
        Email:     user.Email,
        Status:    user.Status.String(),
        CreatedAt: user.CreatedAt,
    }
}
```

### 分页查询模式

**分页参数标准化**:
```go
type PaginationRequest struct {
    Page     int `form:"page,default=1" binding:"min=1"`
    PageSize int `form:"page_size,default=20" binding:"min=1,max=100"`
}

type SortRequest struct {
    Sort  string `form:"sort,default=created_at"`
    Order string `form:"order,default=desc" binding:"oneof=asc desc"`
}

type FilterRequest struct {
    Status    string `form:"status"`
    Keyword   string `form:"q"`
    CreatedAt string `form:"created_after"`
}
```

**分页响应格式**:
```go
type PaginatedResponse struct {
    Items      interface{}    `json:"items"`
    Pagination PaginationInfo `json:"pagination"`
}

type PaginationInfo struct {
    Page       int `json:"page"`
    PageSize   int `json:"page_size"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
}
```

## 🧪 API 测试策略

### 单元测试

**Handler 测试示例**:
```go
func TestUserHandler_GetUser(t *testing.T) {
    tests := []struct {
        name           string
        userID         string
        mockSetup      func(*mocks.UserService)
        expectedStatus int
        expectedCode   int
    }{
        {
            name:   "成功获取用户",
            userID: "user_123",
            mockSetup: func(m *mocks.UserService) {
                user := &models.User{
                    ID:    "user_123",
                    Name:  "测试用户",
                    Email: "test@example.com",
                }
                m.EXPECT().GetUserByID(gomock.Any(), "user_123").
                    Return(user, nil)
            },
            expectedStatus: 200,
            expectedCode:   0,
        },
        {
            name:   "用户不存在",
            userID: "user_999",
            mockSetup: func(m *mocks.UserService) {
                m.EXPECT().GetUserByID(gomock.Any(), "user_999").
                    Return(nil, errors.New("用户不存在"))
            },
            expectedStatus: 404,
            expectedCode:   2001,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试实现
        })
    }
}
```

### 集成测试

**API 集成测试**:
```go
func TestAuthAPI_Integration(t *testing.T) {
    // 设置测试环境
    testApp := setupTestApp(t)
    defer testApp.Cleanup()
    
    t.Run("完整登录流程", func(t *testing.T) {
        // 1. 获取验证码
        captchaResp := testApp.GET("/api/v1/captcha/generate").
            Expect().
            Status(200).
            JSON().Object()
        
        captchaID := captchaResp.Value("data").Object().
            Value("captcha_id").String().Raw()
        
        // 2. 登录请求
        loginReq := map[string]interface{}{
            "email":          "admin@example.com",
            "password":       "admin123",
            "captcha_id":     captchaID,
            "captcha_answer": "test",
        }
        
        loginResp := testApp.POST("/api/v1/auth/login").
            WithJSON(loginReq).
            Expect().
            Status(200).
            JSON().Object()
        
        // 3. 验证响应
        token := loginResp.Value("data").Object().
            Value("access_token").String().NotEmpty().Raw()
        
        // 4. 使用 Token 访问受保护资源
        testApp.GET("/api/v1/users/profile").
            WithHeader("Authorization", "Bearer "+token).
            Expect().
            Status(200)
    })
}
```

## 📊 API 监控与日志

### 请求日志记录

```go
// 中间件记录 API 访问日志
func APILogMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // 记录请求开始
        logger.InfoWithTrace(c.Request.Context(), "API请求开始",
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.String("user_agent", c.GetHeader("User-Agent")),
            zap.String("client_ip", c.ClientIP()),
        )
        
        c.Next()
        
        // 记录请求结束
        duration := time.Since(start)
        logger.InfoWithTrace(c.Request.Context(), "API请求完成",
            zap.Int("status", c.Writer.Status()),
            zap.Duration("duration", duration),
            zap.Int("response_size", c.Writer.Size()),
        )
    }
}
```

### 性能指标收集

```go
// Prometheus 指标定义
var (
    apiRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_requests_total",
            Help: "Total number of API requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    apiRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "api_request_duration_seconds",
            Help: "API request duration in seconds",
        },
        []string{"method", "endpoint"},
    )
)

// 指标收集中间件
func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        
        apiRequestsTotal.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            status,
        ).Inc()
        
        apiRequestDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
        ).Observe(duration)
    }
}
```

## 📚 API 文档管理

### Swagger 文档生成

**标准注释格式**:
```go
// GetUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取租户下的用户列表，支持搜索和过滤
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "用户状态" Enums(active, inactive)
// @Param q query string false "搜索关键词"
// @Security BearerAuth
// @Success 200 {object} response.Response{data=GetUsersResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
    // 实现代码
}
```

### API 文档最佳实践

1. **完整性**: 所有公开 API 都必须有文档
2. **准确性**: 文档与实现保持同步
3. **示例**: 提供完整的请求/响应示例
4. **错误说明**: 详细说明所有可能的错误情况
5. **版本管理**: 记录 API 变更历史

## 🚀 开发工作流

### API 开发标准流程

1. **设计 API**
   - 定义 URL 结构和 HTTP 方法
   - 设计请求/响应格式
   - 确定认证和权限要求

2. **实现 DTO**
   ```go
   // 定义请求和响应结构
   type CreateUserRequest struct { /* ... */ }
   type UserResponse struct { /* ... */ }
   ```

3. **实现 Handler**
   ```go
   func (h *UserHandler) CreateUser(c *gin.Context) {
       // 参数验证 → 服务调用 → 响应构建
   }
   ```

4. **注册路由**
   ```go
   router.POST("/users", handlers.User.CreateUser)
   ```

5. **编写测试**
   ```go
   func TestUserHandler_CreateUser(t *testing.T) { /* ... */ }
   ```

6. **生成文档**
   ```bash
   make docs  # 生成 Swagger 文档
   ```

7. **质量检查**
   ```bash
   make wire && make test
   ```

## 📋 API 检查清单

**开发完成检查**:
- [ ] URL 设计符合 RESTful 规范
- [ ] 请求/响应格式标准化
- [ ] 错误处理完整
- [ ] 认证授权正确实现
- [ ] 多租户支持
- [ ] 输入验证完整
- [ ] 日志记录完整
- [ ] 单元测试覆盖
- [ ] 集成测试通过
- [ ] API 文档完整
- [ ] 性能指标监控

**安全检查**:
- [ ] 敏感数据不在 URL 中传递
- [ ] 输入参数正确验证
- [ ] SQL 注入防护
- [ ] XSS 防护
- [ ] CSRF 防护
- [ ] 速率限制
- [ ] 权限边界检查

## 📖 相关文档

- 🏗️ [架构设计规范](./architecture.md) - 了解分层架构和设计原则
- 🧪 [测试指南](./testing-guide.md) - 学习测试策略和实现方法
- 📋 [认证 API](../api/auth-api.md) - 认证相关 API 详细文档
- 📋 [权限 API](../api/permission-api.md) - 权限管理 API 详细文档

---

**重要提醒**：API 是系统的对外接口，其设计质量直接影响系统的可用性和开发效率。请严格遵循本指南的规范，确保 API 的一致性和专业性。