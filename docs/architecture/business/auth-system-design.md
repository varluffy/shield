# 认证系统设计

## 🎯 设计目标

UltraFit认证系统基于JWT令牌机制，提供安全、高效的用户身份验证和会话管理。支持多租户环境下的邮箱+密码登录方式。

## 🏗️ 架构设计

### 1. 核心组件

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   登录接口      │    │   JWT服务       │    │   用户服务      │
│   LoginHandler  │───▶│   JWTService    │───▶│   UserService   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 2. 认证流程

```
用户登录请求 → 验证邮箱密码 → 生成JWT令牌 → 返回令牌
```

## 🔐 JWT令牌设计

### 1. 访问令牌 (Access Token)

```go
// JWT载荷结构
type JWTClaims struct {
    UserID     string   `json:"user_id"`
    TenantID   string   `json:"tenant_id"`
    Email      string   `json:"email"`
    Name       string   `json:"name"`
    Roles      []string `json:"roles"`
    Permissions []string `json:"permissions"`
    TokenType  string   `json:"token_type"` // "access"
    jwt.RegisteredClaims
}

// 令牌配置
const (
    AccessTokenExpiry  = 2 * time.Hour    // 访问令牌有效期
    RefreshTokenExpiry = 30 * 24 * time.Hour // 刷新令牌有效期
)
```

### 2. 刷新令牌 (Refresh Token)

```go
// 刷新令牌载荷
type RefreshTokenClaims struct {
    UserID    string `json:"user_id"`
    TenantID  string `json:"tenant_id"`
    TokenID   string `json:"token_id"`   // 唯一标识
    TokenType string `json:"token_type"` // "refresh"
    jwt.RegisteredClaims
}

// 令牌存储模型
type RefreshToken struct {
    ID        string    `gorm:"primary_key;type:varchar(36)" json:"id"`
    UserID    string    `gorm:"type:varchar(36);not null;index" json:"user_id"`
    TenantID  string    `gorm:"type:varchar(36);not null;index" json:"tenant_id"`
    TokenHash string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"token_hash"`
    ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
    IsRevoked bool      `gorm:"default:false" json:"is_revoked"`
    UserAgent string    `gorm:"type:text" json:"user_agent"`
    IPAddress string    `gorm:"type:varchar(45)" json:"ip_address"`
}
```

## 🔧 核心服务实现

### 1. JWT服务

```go
// JWT服务接口
type JWTService interface {
    GenerateTokenPair(ctx context.Context, user *User) (*TokenPair, error)
    ValidateAccessToken(ctx context.Context, tokenString string) (*JWTClaims, error)
    RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error)
    RevokeRefreshToken(ctx context.Context, tokenID string) error
    RevokeUserTokens(ctx context.Context, userID string) error
}

// 令牌对
type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    TokenType    string `json:"token_type"`
    ExpiresIn    int64  `json:"expires_in"`
}
```

### 2. 认证服务

```go
// 认证服务接口
type AuthService interface {
    Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
    Logout(ctx context.Context, req *LogoutRequest) error
    RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error)
    ChangePassword(ctx context.Context, req *ChangePasswordRequest) error
}

// 登录请求
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    TenantID string `json:"tenant_id" binding:"required"`
}

// 登录响应
type LoginResponse struct {
    User         *UserInfo  `json:"user"`
    TokenPair    *TokenPair `json:"tokens"`
    LastLoginAt  *time.Time `json:"last_login_at"`
}
```

### 3. 登录安全控制

```go
// 登录尝试记录
type LoginAttempt struct {
    ID        string    `gorm:"primary_key;type:varchar(36)" json:"id"`
    UserID    string    `gorm:"type:varchar(36);index" json:"user_id"`
    TenantID  string    `gorm:"type:varchar(36);index" json:"tenant_id"`
    Email     string    `gorm:"type:varchar(255);index" json:"email"`
    IPAddress string    `gorm:"type:varchar(45)" json:"ip_address"`
    Success   bool      `gorm:"default:false" json:"success"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// 登录失败5次锁定30分钟
const (
    MaxLoginAttempts = 5
    LockoutDuration  = 30 * time.Minute
)
```

## 🔒 认证中间件

```go
// 认证中间件
func AuthMiddleware(jwtService JWTService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 提取令牌
        tokenString := extractToken(c)
        if tokenString == "" {
            c.JSON(401, gin.H{"error": "缺少访问令牌"})
            c.Abort()
            return
        }
        
        // 2. 验证令牌
        claims, err := jwtService.ValidateAccessToken(c.Request.Context(), tokenString)
        if err != nil {
            c.JSON(401, gin.H{"error": "无效的访问令牌"})
            c.Abort()
            return
        }
        
        // 3. 设置用户上下文
        c.Set("user_id", claims.UserID)
        c.Set("tenant_id", claims.TenantID)
        c.Set("user_email", claims.Email)
        c.Set("user_roles", claims.Roles)
        c.Set("user_permissions", claims.Permissions)
        
        c.Next()
    }
}

// 提取令牌
func extractToken(c *gin.Context) string {
    authHeader := c.GetHeader("Authorization")
    if authHeader != "" {
        parts := strings.Split(authHeader, " ")
        if len(parts) == 2 && parts[0] == "Bearer" {
            return parts[1]
        }
    }
    return c.Query("access_token")
}
```

## 📊 密码安全

### 1. 密码哈希

```go
// 密码服务
type PasswordService interface {
    HashPassword(password string) (string, error)
    VerifyPassword(password, hash string) bool
    ValidatePasswordStrength(password string) error
}

// 密码强度验证
func (s *passwordService) ValidatePasswordStrength(password string) error {
    if len(password) < 8 {
        return ErrPasswordTooShort
    }
    
    var hasUpper, hasLower, hasNumber bool
    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsNumber(char):
            hasNumber = true
        }
    }
    
    if !hasUpper || !hasLower || !hasNumber {
        return ErrPasswordTooWeak
    }
    
    return nil
}
```

### 2. 密码策略

```go
// 密码策略配置
type PasswordPolicy struct {
    MinLength        int           `json:"min_length"`        // 最小长度：8
    RequireUppercase bool          `json:"require_uppercase"` // 需要大写字母
    RequireLowercase bool          `json:"require_lowercase"` // 需要小写字母
    RequireNumbers   bool          `json:"require_numbers"`   // 需要数字
    MaxAge           time.Duration `json:"max_age"`           // 密码过期时间
    PreventReuse     int           `json:"prevent_reuse"`     // 防止重复使用
}
```

## 🔄 令牌刷新机制

```go
// 令牌刷新流程
func (s *authService) RefreshToken(ctx context.Context, req *RefreshTokenRequest) (*RefreshTokenResponse, error) {
    // 1. 验证刷新令牌
    claims, err := s.jwtService.ValidateRefreshToken(ctx, req.RefreshToken)
    if err != nil {
        return nil, ErrInvalidRefreshToken
    }
    
    // 2. 检查令牌是否被撤销
    if isRevoked, _ := s.jwtService.IsTokenRevoked(ctx, claims.TokenID); isRevoked {
        return nil, ErrTokenRevoked
    }
    
    // 3. 获取用户信息
    user, err := s.userService.GetUser(ctx, claims.UserID)
    if err != nil {
        return nil, err
    }
    
    // 4. 生成新的令牌对
    tokenPair, err := s.jwtService.GenerateTokenPair(ctx, user)
    if err != nil {
        return nil, err
    }
    
    // 5. 撤销旧的刷新令牌
    s.jwtService.RevokeRefreshToken(ctx, claims.TokenID)
    
    return &RefreshTokenResponse{
        TokenPair: tokenPair,
        User:      convertToUserInfo(user),
    }, nil
}
```

## 📈 监控和审计

### 1. 认证事件记录

```go
// 认证事件
type AuthEvent struct {
    ID        string    `gorm:"primary_key;type:varchar(36)" json:"id"`
    UserID    string    `gorm:"type:varchar(36);index" json:"user_id"`
    TenantID  string    `gorm:"type:varchar(36);index" json:"tenant_id"`
    Event     string    `gorm:"type:varchar(50);index" json:"event"`
    IPAddress string    `gorm:"type:varchar(45)" json:"ip_address"`
    UserAgent string    `gorm:"type:text" json:"user_agent"`
    Success   bool      `gorm:"default:true" json:"success"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// 事件类型
const (
    EventLogin          = "login"
    EventLogout         = "logout"
    EventTokenRefresh   = "token_refresh"
    EventPasswordChange = "password_change"
    EventAccountLocked  = "account_locked"
)
```

## 🎯 API接口设计

### 1. 登录接口

```http
POST /api/v1/auth/login
Content-Type: application/json

{
    "email": "user@example.com",
    "password": "Password123",
    "tenant_id": "tenant1"
}

# 响应
{
    "user": {
        "id": "user-id",
        "email": "user@example.com",
        "name": "用户名",
        "roles": ["user"]
    },
    "tokens": {
        "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
        "token_type": "Bearer",
        "expires_in": 7200
    }
}
```

### 2. 令牌刷新接口

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
    "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}

# 响应
{
    "tokens": {
        "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
        "token_type": "Bearer",
        "expires_in": 7200
    }
}
```

### 3. 登出接口

```http
POST /api/v1/auth/logout
Authorization: Bearer <access_token>

# 响应
{
    "message": "登出成功"
}
```

## 🎯 最佳实践

### 1. 安全建议
- 使用RSA-256算法签名JWT令牌
- 访问令牌有效期设置为2小时
- 刷新令牌有效期设置为30天
- 密码使用bcrypt哈希，成本因子≥12
- 记录所有认证事件用于审计

### 2. 性能优化
- 使用Redis缓存用户权限信息
- 定期清理过期的令牌记录
- 使用连接池管理数据库连接

### 3. 错误处理
- 统一的错误响应格式
- 不泄露敏感信息的错误消息
- 详细的错误日志记录

## 🎯 总结

UltraFit认证系统提供了完整的身份验证解决方案：

1. **JWT令牌管理**: 访问令牌和刷新令牌的完整生命周期管理
2. **多租户支持**: 令牌中包含租户信息，支持租户隔离
3. **安全防护**: 密码强度验证、登录失败锁定、令牌撤销等
4. **监控审计**: 完整的认证事件记录和异常监控
5. **高性能**: 基于JWT的无状态认证，支持水平扩展

该系统为后续的权限控制和业务功能提供了安全可靠的身份验证基础。 