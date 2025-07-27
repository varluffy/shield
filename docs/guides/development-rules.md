# UltraFit 开发规则与约束

## 📋 文档概述

本文档定义了 UltraFit 项目的核心开发规则和约束，确保代码质量、架构一致性和团队协作效率。

**这些规则是我们代码质量的基石，请认真遵循！** 🏗️

## 🏗️ 架构约束

### 1. 分层架构强制规则

#### 1.1 分层定义
- **Handler层**：HTTP请求处理，参数绑定和响应格式化
- **Service层**：业务逻辑处理，事务管理
- **Repository层**：数据访问，数据库操作
- **Model层**：数据模型定义

#### 1.2 依赖规则
- Handler 只能调用 Service，不能直接调用 Repository
- Service 可以调用 Repository 和其他 Service
- Repository 只能处理数据访问，不能包含业务逻辑
- 所有跨层调用必须通过接口

#### 1.3 违规示例
```go
// ❌ 错误：Handler直接调用Repository
func (h *UserHandler) CreateUser(c *gin.Context) {
    user := &models.User{}
    h.userRepo.Create(user) // 违规！
}

// ✅ 正确：Handler通过Service调用
func (h *UserHandler) CreateUser(c *gin.Context) {
    user := &models.User{}
    err := h.userService.CreateUser(ctx, user)
}
```

### 2. 接口驱动开发规则

#### 2.1 接口定义要求
- 所有Service必须定义接口
- 所有Repository必须定义接口
- 接口应该放在对应的包中

#### 2.2 依赖注入规则
- 通过Wire进行依赖注入
- Provider必须返回接口类型
- 避免直接依赖具体实现

#### 2.3 接口示例
```go
// 定义接口
type UserService interface {
    CreateUser(ctx context.Context, user *models.User) error
    GetUserByID(ctx context.Context, id uint) (*models.User, error)
}

// Wire Provider返回接口
func NewUserService(repo repositories.UserRepository) services.UserService {
    return &userService{repo: repo}
}
```

### 3. 验证码系统架构规则

#### 3.1 验证码服务规则
- 验证码服务必须使用接口定义
- 必须使用Redis作为分布式存储
- 验证码类型必须可配置
- 验证码过期时间必须可配置

#### 3.2 验证码集成规则
- 登录接口必须验证图形验证码
- 验证码验证失败必须记录日志
- 验证码必须一次性使用
- 验证码ID和答案不能为空

#### 3.3 验证码接口规范
```go
// 验证码服务接口
type CaptchaService interface {
    GenerateCaptcha() (id, b64s string, err error)
    VerifyCaptcha(id, answer string) error
}

// 验证码存储接口
type CaptchaStore interface {
    Set(id string, value string)
    Get(id string, clear bool) string
    Verify(id, answer string, clear bool) bool
}
```

## 🎯 代码规范

### 1. 命名规范

#### 1.1 变量命名
- 使用驼峰命名法
- 布尔值使用 `is`、`has`、`can` 前缀
- 常量使用全大写加下划线

#### 1.2 函数命名
- 使用动词开头
- 返回布尔值的函数使用 `is`、`has`、`can` 前缀
- 验证码相关函数使用 `Captcha` 前缀

#### 1.3 接口命名
- 使用名词结尾
- 单一职责的接口使用 `-er` 后缀
- 验证码接口使用 `CaptchaService`、`CaptchaStore` 命名

### 2. 错误处理规范

#### 2.1 错误返回规则
- 所有可能失败的函数必须返回error
- 使用pkg/errors包装错误
- 错误信息必须包含上下文

#### 2.2 错误处理示例
```go
// ✅ 正确的错误处理
func (s *userService) CreateUser(ctx context.Context, user *models.User) error {
    if err := s.validateUser(user); err != nil {
        return errors.Wrap(err, "用户验证失败")
    }
    
    if err := s.userRepo.Create(ctx, user); err != nil {
        return errors.Wrap(err, "创建用户失败")
    }
    
    return nil
}

// 验证码错误处理
func (s *captchaService) VerifyCaptcha(id, answer string) error {
    if id == "" || answer == "" {
        return errors.New("验证码ID或答案不能为空")
    }
    
    if !s.store.Verify(id, answer, true) {
        return errors.New("验证码错误或已过期")
    }
    
    return nil
}
```

### 3. 日志记录规范

#### 3.1 日志级别
- **Debug**：调试信息
- **Info**：一般信息
- **Warn**：警告信息
- **Error**：错误信息

#### 3.2 日志格式
- 使用结构化日志（zap）
- 包含trace_id用于链路追踪
- 记录关键业务操作

#### 3.3 验证码日志规范
```go
// 记录验证码操作
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

### 4. 配置管理规范

#### 4.1 配置结构
- 使用结构化配置
- 支持多环境配置
- 敏感信息使用环境变量

#### 4.2 验证码配置
```go
type CaptchaConfig struct {
    Enabled     bool          `mapstructure:"enabled"`
    Type        string        `mapstructure:"type"`
    Width       int           `mapstructure:"width"`
    Height      int           `mapstructure:"height"`
    Length      int           `mapstructure:"length"`
    NoiseCount  int           `mapstructure:"noise_count"`
    Expiration  time.Duration `mapstructure:"expiration"`
}
```

## 📋 开发流程约束

### 1. 需求驱动开发

#### 1.1 变更原则
- 只针对明确的需求进行修改
- 不进行预测性开发
- 变更必须有可验证的价值

#### 1.2 变更范围限制
- 单次变更专注单一需求
- 避免同时修改多个模块
- 重构和功能开发分离

### 2. 代码审查规则

#### 2.1 审查级别
- **低风险**：单元测试、文档更新 → 1人审查
- **中风险**：业务逻辑修改 → 2人审查
- **高风险**：架构变更、安全相关 → 3人审查

#### 2.2 审查重点
- 架构一致性
- 代码质量
- 安全性
- 性能影响

### 3. 测试策略

#### 3.1 测试覆盖要求
- 单元测试覆盖率 > 80%
- 核心业务逻辑必须有测试
- 接口层必须有集成测试

#### 3.2 验证码测试要求
- 验证码生成功能测试
- 验证码验证功能测试
- Redis存储功能测试
- 分布式场景测试

## 🔍 代码审查检查清单

### 1. Handler层检查
- [ ] 是否只处理HTTP层逻辑，不包含业务逻辑？
- [ ] 是否正确使用参数绑定和验证？
- [ ] 是否实现了统一的错误处理？
- [ ] 是否将Context正确传递给Service层？
- [ ] 验证码相关接口是否正确验证？

### 2. Service层检查
- [ ] 是否实现了接口定义？
- [ ] 是否正确处理业务逻辑和错误？
- [ ] 是否使用Context进行超时控制？
- [ ] 是否正确使用事务处理？
- [ ] 验证码服务是否正确集成？

### 3. Repository层检查
- [ ] 是否使用Repository模式封装数据访问？
- [ ] 是否避免了N+1查询问题？
- [ ] 是否使用了参数化查询防止SQL注入？
- [ ] 是否正确处理了GORM错误？

### 4. Wire配置检查
- [ ] 是否返回接口类型而非具体实现？
- [ ] 是否正确绑定接口到实现？
- [ ] 是否避免了循环依赖？
- [ ] 是否为测试创建了专门的Provider？
- [ ] 验证码服务是否正确注入？

### 5. 验证码系统检查
- [ ] 是否使用了Redis作为分布式存储？
- [ ] 是否正确实现了验证码生成和验证？
- [ ] 是否设置了合理的过期时间？
- [ ] 是否在登录流程中集成了验证码验证？
- [ ] 是否记录了验证码相关的操作日志？

### 6. 配置管理检查
- [ ] 是否使用结构化配置绑定？
- [ ] 是否为所有配置项设置了默认值？
- [ ] 是否正确处理了敏感信息？
- [ ] 是否实现了配置验证？
- [ ] 验证码配置是否完整？

### 7. 安全检查
- [ ] 是否遵循了分层架构，没有跨层直接访问？
- [ ] 是否正确实现了认证和授权？
- [ ] 是否使用了参数化查询？
- [ ] 是否正确处理了用户输入验证？
- [ ] 验证码是否具有足够的安全强度？

### 8. 性能检查
- [ ] 是否避免了不必要的数据库查询？
- [ ] 是否正确使用了缓存策略？
- [ ] 是否避免了内存泄漏？
- [ ] 是否进行了适当的并发控制？
- [ ] 验证码系统是否影响登录性能？

## 🛡️ 安全规范

### 1. 数据安全
- 敏感数据必须加密存储
- 数据库连接必须使用SSL
- 用户密码必须使用bcrypt加密
- 验证码答案必须安全存储

### 2. 认证安全
- 使用JWT进行身份验证
- 实现refresh token机制
- 登录失败次数限制
- 图形验证码防护

### 3. 权限安全
- 实现基于角色的访问控制
- API访问权限验证
- 租户数据隔离
- 权限变更审计

### 4. 验证码安全
- 验证码必须一次性使用
- 设置合理的过期时间
- 防止暴力破解
- 验证码复杂度控制

## 🚀 性能规范

### 1. 数据库性能
- 合理使用索引
- 避免N+1查询
- 使用数据库连接池
- 实现查询缓存

### 2. 缓存策略
- 使用Redis缓存热点数据
- 实现缓存预热
- 合理设置缓存过期时间
- 验证码缓存优化

### 3. 并发控制
- 使用goroutine池
- 实现请求限流
- 数据库事务控制
- 避免死锁

## 📊 监控和日志

### 1. 监控指标
- 请求响应时间
- 错误率统计
- 数据库连接数
- 验证码生成/验证成功率

### 2. 日志规范
- 使用结构化日志
- 记录关键业务操作
- 包含trace_id
- 验证码操作日志

### 3. 告警机制
- 错误率超阈值告警
- 系统资源告警
- 安全事件告警
- 验证码异常告警

## 📚 文档规范

### 1. 代码文档
- 公开函数必须有注释
- 复杂逻辑必须有说明
- 接口文档必须完整
- 验证码相关函数必须有文档

### 2. API文档
- 使用OpenAPI规范
- 包含请求/响应示例
- 错误码说明
- 验证码接口文档

### 3. 架构文档
- 设计决策记录
- 架构图更新
- 依赖关系说明
- 验证码系统设计文档

---

## 📋 总结

遵循这些开发规则和约束，确保 UltraFit 项目的代码质量、架构一致性和团队协作效率。所有开发人员都应该熟悉并严格遵守这些规范。

验证码系统的集成为项目增加了重要的安全保护，开发过程中必须严格按照规范进行实现和测试。 