# UltraFit 验证码系统架构设计

## 1. 系统概述

UltraFit 验证码系统基于 [base64Captcha](https://github.com/mojocn/base64Captcha) 包实现，使用 Redis 作为分布式存储，为登录安全提供图形验证码保护。

## 2. 技术选型

### 2.1 核心组件
- **验证码库**：[mojocn/base64Captcha](https://github.com/mojocn/base64Captcha) v1.2.2
- **存储方案**：Redis（替代默认内存存储）
- **验证码类型**：数字、字符串、数学运算
- **编码格式**：Base64 图片字符串

### 2.2 技术优势
- **分布式支持**：Redis 存储支持多实例共享
- **多种类型**：支持数字、字符、数学等验证码
- **自定义能力**：可定制验证码样式和难度
- **高性能**：Base64 编码，前端直接展示

## 3. 架构设计

### 3.1 组件关系图
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │    │     Redis       │
│                 │    │                 │    │                 │
│  1. 请求验证码   │───▶│  2. 生成验证码   │    │                 │
│  2. 显示图片     │◀───│  3. 返回Base64  │    │                 │
│  3. 用户输入     │    │                 │    │                 │
│  4. 提交登录     │───▶│  4. 验证验证码   │───▶│  5. 存储/验证    │
│                 │    │  5. 处理登录     │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 3.2 数据流程
```
1. 获取验证码：
   Frontend → Backend → base64Captcha → Redis → Backend → Frontend

2. 验证流程：
   Frontend → Backend → Redis → Backend → 登录逻辑
```

## 4. 核心实现

### 4.1 Redis Store 实现

```go
package captcha

import (
    "context"
    "time"
    
    "github.com/go-redis/redis/v8"
    "github.com/mojocn/base64Captcha"
)

// RedisCaptchaStore Redis 验证码存储
type RedisCaptchaStore struct {
    redisClient *redis.Client
    prefix      string
    expiration  time.Duration
}

// NewRedisCaptchaStore 创建Redis验证码存储
func NewRedisCaptchaStore(redisClient *redis.Client) base64Captcha.Store {
    return &RedisCaptchaStore{
        redisClient: redisClient,
        prefix:      "captcha:",
        expiration:  5 * time.Minute, // 验证码5分钟过期
    }
}

// Set 存储验证码答案
func (r *RedisCaptchaStore) Set(id string, value string) {
    key := r.prefix + id
    ctx := context.Background()
    r.redisClient.Set(ctx, key, value, r.expiration)
}

// Get 获取验证码答案
func (r *RedisCaptchaStore) Get(id string, clear bool) string {
    key := r.prefix + id
    ctx := context.Background()
    
    result := r.redisClient.Get(ctx, key)
    if result.Err() != nil {
        return ""
    }
    
    value := result.Val()
    
    // 如果需要清除，删除Redis中的记录
    if clear {
        r.redisClient.Del(ctx, key)
    }
    
    return value
}

// Verify 验证验证码
func (r *RedisCaptchaStore) Verify(id, answer string, clear bool) bool {
    return r.Get(id, clear) == answer
}
```

### 4.2 验证码服务

```go
package captcha

import (
    "errors"
    
    "github.com/go-redis/redis/v8"
    "github.com/mojocn/base64Captcha"
)

// CaptchaService 验证码服务
type CaptchaService struct {
    captcha *base64Captcha.Captcha
    store   base64Captcha.Store
}

// CaptchaConfig 验证码配置
type CaptchaConfig struct {
    Width      int    `mapstructure:"width"`
    Height     int    `mapstructure:"height"`
    Length     int    `mapstructure:"length"`
    Source     string `mapstructure:"source"`
    NoiseCount int    `mapstructure:"noise_count"`
}

// NewCaptchaService 创建验证码服务
func NewCaptchaService(redisClient *redis.Client, config *CaptchaConfig) *CaptchaService {
    store := NewRedisCaptchaStore(redisClient)
    
    driver := &base64Captcha.DriverDigit{
        Height:   config.Height,
        Width:    config.Width,
        Length:   config.Length,
        MaxSkew:  0.7,
        DotCount: config.NoiseCount,
    }
    
    captcha := base64Captcha.NewCaptcha(driver, store)
    
    return &CaptchaService{
        captcha: captcha,
        store:   store,
    }
}

// GenerateCaptcha 生成验证码
func (s *CaptchaService) GenerateCaptcha() (id, b64s string, err error) {
    return s.captcha.Generate()
}

// VerifyCaptcha 验证验证码
func (s *CaptchaService) VerifyCaptcha(id, answer string) error {
    if id == "" || answer == "" {
        return errors.New("验证码ID或答案不能为空")
    }
    
    if !s.store.Verify(id, answer, true) {
        return errors.New("验证码错误或已过期")
    }
    
    return nil
}
```

## 5. API 接口设计

### 5.1 获取验证码接口

**GET** `/api/v1/captcha/generate`

**响应数据**:
```json
{
  "code": 200,
  "message": "验证码生成成功",
  "data": {
    "captcha_id": "bp8RkzOTBEObGLvueygk",
    "captcha_image": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAJYAAAA..."
  }
}
```

### 5.2 更新登录接口

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
  "code": 200,
  "message": "登录成功",
  "data": {
    "user": { ... },
    "tokens": { ... },
    "accessible_tenants": [ ... ]
  }
}
```

## 6. 前端集成

### 6.1 验证码组件示例

```javascript
const CaptchaComponent = {
  data() {
    return {
      captchaId: '',
      captchaImage: '',
      captchaAnswer: '',
      loading: false
    }
  },
  
  mounted() {
    this.refreshCaptcha()
  },
  
  methods: {
    async refreshCaptcha() {
      this.loading = true
      try {
        const response = await axios.get('/api/v1/captcha/generate')
        this.captchaId = response.data.data.captcha_id
        this.captchaImage = response.data.data.captcha_image
        this.captchaAnswer = ''
      } catch (error) {
        console.error('获取验证码失败:', error)
      } finally {
        this.loading = false
      }
    },
    
    async login() {
      const loginData = {
        email: this.email,
        password: this.password,
        captcha_id: this.captchaId,
        captcha_answer: this.captchaAnswer
      }
      
      try {
        const response = await axios.post('/api/v1/auth/login', loginData)
        // 处理登录成功
      } catch (error) {
        // 登录失败，刷新验证码
        this.refreshCaptcha()
      }
    }
  }
}
```

## 7. 配置管理

### 7.1 配置文件

```yaml
# config.yaml
captcha:
  enabled: true          # 是否启用验证码
  type: "digit"         # 验证码类型：digit, string, math
  width: 160            # 图片宽度
  height: 60            # 图片高度
  length: 4             # 验证码长度
  source: "1234567890"  # 字符源
  noise_count: 5        # 噪点数量
  expiration: "5m"      # 过期时间
```

### 7.2 配置结构体

```go
type CaptchaConfig struct {
    Enabled     bool          `mapstructure:"enabled"`
    Type        string        `mapstructure:"type"`
    Width       int           `mapstructure:"width"`
    Height      int           `mapstructure:"height"`
    Length      int           `mapstructure:"length"`
    Source      string        `mapstructure:"source"`
    NoiseCount  int           `mapstructure:"noise_count"`
    Expiration  time.Duration `mapstructure:"expiration"`
}
```

## 8. 安全考虑

### 8.1 安全策略
- **有效期控制**：验证码5分钟过期
- **一次性使用**：验证后立即删除
- **防暴力破解**：结合登录失败次数限制
- **图片复杂度**：适当的噪点和扭曲

### 8.2 攻击防护
- **重放攻击**：验证码ID唯一性保护
- **暴力破解**：验证码复杂度 + 失败次数限制
- **存储安全**：Redis 密码保护
- **网络安全**：HTTPS 传输保护

## 9. 监控和运维

### 9.1 关键指标
- **生成成功率**：验证码生成成功率
- **验证成功率**：用户验证成功率
- **过期清理**：Redis 过期键清理
- **性能指标**：生成和验证耗时

### 9.2 日志记录
```go
logger.InfoWithTrace(ctx, "验证码生成成功",
    zap.String("captcha_id", id),
    zap.String("captcha_type", "digit"),
    zap.String("ip_address", clientIP),
)

logger.InfoWithTrace(ctx, "验证码验证",
    zap.String("captcha_id", id),
    zap.Bool("verify_result", success),
    zap.String("ip_address", clientIP),
)
```

---

## 总结

UltraFit 验证码系统通过集成 [base64Captcha](https://github.com/mojocn/base64Captcha) 和 Redis 存储，提供了一个安全、高性能、分布式的验证码解决方案。该系统支持多种验证码类型，具备良好的扩展性和可维护性，能够有效防止恶意登录攻击。 