# Go 数据库使用指南

UltraFit项目的数据层技术栈整合指南，包括GORM、Redis和数据库设计规范。

## 🎯 GORM 核心原则

### Repository模式
- 使用**Repository模式**封装GORM操作，避免在service层直接使用GORM
- 定义清晰的数据访问接口，与具体ORM实现解耦
- 实施**事务处理**包装复合操作，确保数据一致性

### 模型定义规范
```go
// 基础模型（带UUID）
type BaseModel struct {
    ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
    UUID      string         `gorm:"type:char(36);not null;uniqueIndex" json:"uuid"`
    CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// 租户模型
type TenantModel struct {
    BaseModel
    TenantID  uint64 `gorm:"not null;index" json:"tenant_id"`
}

// 业务模型示例
type User struct {
    TenantModel
    Email                string     `gorm:"type:varchar(255);not null;uniqueIndex:uk_tenant_email" json:"email"`
    Password             string     `gorm:"type:varchar(255);not null" json:"-"`
    Name                 string     `gorm:"type:varchar(100)" json:"name"`
    Status               string     `gorm:"type:varchar(20);default:'active'" json:"status"`
    LastLoginAt          *time.Time `json:"last_login_at"`
}

func (User) TableName() string {
    return "users"
}
```

## 🏗️ Repository 接口设计

### 标准Repository接口
```go
type UserRepository interface {
    // 基础CRUD
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id uint64) (*User, error)
    GetByUUID(ctx context.Context, uuid string) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id uint64) error
    
    // 查询接口
    List(ctx context.Context, filter UserFilter) ([]*User, int64, error)
    GetByEmailAndTenant(ctx context.Context, email string, tenantID uint64) (*User, error)
    
    // 事务支持
    WithTx(tx *gorm.DB) UserRepository
}

type UserFilter struct {
    TenantID uint64
    Name     string
    Email    string
    Status   string
    Page     int
    Limit    int
}
```

### Repository实现示例
```go
type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByUUID(ctx context.Context, uuid string) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, ErrUserNotFound()
    }
    return &user, err
}

func (r *userRepository) List(ctx context.Context, filter UserFilter) ([]*User, int64, error) {
    var users []*User
    var total int64
    
    query := r.db.WithContext(ctx).Model(&User{})
    
    // 多租户过滤
    if filter.TenantID > 0 {
        query = query.Where("tenant_id = ?", filter.TenantID)
    }
    
    // 条件过滤
    if filter.Name != "" {
        query = query.Where("name LIKE ?", "%"+filter.Name+"%")
    }
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }
    
    // 获取总数
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // 分页查询
    offset := (filter.Page - 1) * filter.Limit
    err := query.Offset(offset).Limit(filter.Limit).
        Order("created_at DESC").Find(&users).Error
    
    return users, total, err
}

func (r *userRepository) WithTx(tx *gorm.DB) UserRepository {
    return &userRepository{db: tx}
}
```

## 🔄 事务管理

### 事务管理器接口
```go
type TransactionManager interface {
    WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type transactionManager struct {
    db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) TransactionManager {
    return &transactionManager{db: db}
}

func (tm *transactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
    return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 将事务放入上下文
        txCtx := context.WithValue(ctx, "tx", tx)
        return fn(txCtx)
    })
}
```

### 在Service中使用事务
```go
func (s *UserService) CreateUserWithProfile(ctx context.Context, req CreateUserWithProfileRequest) error {
    return s.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
        // 获取事务
        tx := txCtx.Value("tx").(*gorm.DB)
        
        // 使用事务创建用户
        user := &User{...}
        if err := s.userRepo.WithTx(tx).Create(txCtx, user); err != nil {
            return err
        }
        
        // 使用事务创建用户资料
        profile := &UserProfile{UserID: user.ID, ...}
        if err := s.profileRepo.WithTx(tx).Create(txCtx, profile); err != nil {
            return err
        }
        
        return nil
    })
}
```

## 📦 Redis 缓存使用

### Redis配置
```yaml
# configs/config.dev.yaml
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5
  dial_timeout: "5s"
  read_timeout: "3s"
  write_timeout: "3s"
  key_prefix: "shield:dev:"  # 环境前缀
```

### Redis客户端封装
```go
type RedisClient struct {
    client    redis.Cmdable
    keyPrefix string
}

func NewRedisClient(config *RedisConfig) (*RedisClient, error) {
    rdb := redis.NewClient(&redis.Options{
        Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
        Password:     config.Password,
        DB:           config.DB,
        PoolSize:     config.PoolSize,
        MinIdleConns: config.MinIdleConns,
        DialTimeout:  config.DialTimeout,
        ReadTimeout:  config.ReadTimeout,
        WriteTimeout: config.WriteTimeout,
    })
    
    // 测试连接
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := rdb.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to redis: %w", err)
    }
    
    return &RedisClient{
        client:    rdb,
        keyPrefix: config.KeyPrefix,
    }, nil
}

func (r *RedisClient) buildKey(key string) string {
    return r.keyPrefix + key
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    return r.client.Set(ctx, r.buildKey(key), value, expiration).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
    return r.client.Get(ctx, r.buildKey(key)).Result()
}
```

### 验证码存储示例
```go
type CaptchaStore struct {
    redis *RedisClient
}

func NewCaptchaStore(redis *RedisClient) *CaptchaStore {
    return &CaptchaStore{redis: redis}
}

func (s *CaptchaStore) Set(id string, digits []byte) {
    ctx := context.Background()
    key := fmt.Sprintf("captcha:%s", id)
    
    // 验证码有效期5分钟
    s.redis.Set(ctx, key, string(digits), 5*time.Minute)
}

func (s *CaptchaStore) Get(id string, clear bool) []byte {
    ctx := context.Background()
    key := fmt.Sprintf("captcha:%s", id)
    
    result, err := s.redis.Get(ctx, key)
    if err != nil {
        return nil
    }
    
    if clear {
        s.redis.Del(ctx, key)
    }
    
    return []byte(result)
}
```

## 🎯 数据库设计规范

### 命名规范
```sql
-- 表名：小写，下划线分隔，复数形式
CREATE TABLE users (...);
CREATE TABLE user_profiles (...);
CREATE TABLE role_permissions (...);

-- 字段名：小写，下划线分隔
user_id, created_at, email_verified_at

-- 索引名：表名_字段名_类型
idx_users_email          -- 普通索引
uk_users_email           -- 唯一索引
uk_tenant_email          -- 复合唯一索引
```

### 字段类型规范
```sql
-- ID字段：bigint unsigned auto_increment
id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT

-- UUID字段：char(36)
uuid CHAR(36) NOT NULL UNIQUE

-- 租户ID：bigint unsigned
tenant_id BIGINT UNSIGNED NOT NULL

-- 时间字段：timestamp
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
deleted_at TIMESTAMP NULL

-- 字符串字段：varchar(length)
name VARCHAR(100) NOT NULL
email VARCHAR(255) NOT NULL
status VARCHAR(20) DEFAULT 'active'

-- 文本字段：text
description TEXT
content LONGTEXT

-- 布尔字段：tinyint(1)
is_active TINYINT(1) DEFAULT 1
is_deleted TINYINT(1) DEFAULT 0
```

### 索引设计原则
```sql
-- 主键索引
PRIMARY KEY (id)

-- 唯一索引
UNIQUE KEY uk_users_uuid (uuid)
UNIQUE KEY uk_users_tenant_email (tenant_id, email)

-- 普通索引
KEY idx_users_tenant_id (tenant_id)
KEY idx_users_status (status)
KEY idx_users_created_at (created_at)

-- 复合索引（最左前缀原则）
KEY idx_users_tenant_status (tenant_id, status)
KEY idx_users_tenant_created (tenant_id, created_at)

-- 软删除索引
KEY idx_users_deleted_at (deleted_at)
```

## 📊 性能优化

### 查询优化
```go
// ✅ 好的做法：使用索引字段查询
func (r *userRepository) GetByTenantAndEmail(ctx context.Context, tenantID uint64, email string) (*User, error) {
    var user User
    // 利用复合索引 uk_users_tenant_email
    err := r.db.WithContext(ctx).
        Where("tenant_id = ? AND email = ?", tenantID, email).
        First(&user).Error
    return &user, err
}

// ✅ 分页查询优化
func (r *userRepository) ListWithCursor(ctx context.Context, cursor uint64, limit int) ([]*User, error) {
    var users []*User
    query := r.db.WithContext(ctx).Model(&User{})
    
    if cursor > 0 {
        query = query.Where("id > ?", cursor)
    }
    
    err := query.Order("id ASC").Limit(limit).Find(&users).Error
    return users, err
}

// ❌ 避免：不使用索引的查询
func (r *userRepository) SearchByName(ctx context.Context, name string) ([]*User, error) {
    var users []*User
    // LIKE查询无法使用索引，性能差
    err := r.db.WithContext(ctx).
        Where("name LIKE ?", "%"+name+"%").
        Find(&users).Error
    return users, err
}
```

### 预加载优化
```go
// ✅ 使用Preload避免N+1问题
func (r *userRepository) GetUsersWithRoles(ctx context.Context) ([]*User, error) {
    var users []*User
    err := r.db.WithContext(ctx).
        Preload("Roles").
        Find(&users).Error
    return users, err
}

// ✅ 选择性字段查询
func (r *userRepository) GetUserSummary(ctx context.Context, id uint64) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).
        Select("id", "uuid", "name", "email", "status").
        Where("id = ?", id).
        First(&user).Error
    return &user, err
}
```

### 缓存策略
```go
// 用户信息缓存
func (s *UserService) GetUserByUUID(ctx context.Context, uuid string) (*User, error) {
    // 1. 尝试从缓存获取
    cacheKey := fmt.Sprintf("user:uuid:%s", uuid)
    if cached, err := s.redis.Get(ctx, cacheKey); err == nil {
        var user User
        if err := json.Unmarshal([]byte(cached), &user); err == nil {
            return &user, nil
        }
    }
    
    // 2. 从数据库获取
    user, err := s.userRepo.GetByUUID(ctx, uuid)
    if err != nil {
        return nil, err
    }
    
    // 3. 写入缓存
    if data, err := json.Marshal(user); err == nil {
        s.redis.Set(ctx, cacheKey, string(data), 10*time.Minute)
    }
    
    return user, nil
}
```

## 🔍 监控和调试

### 慢查询日志
```go
// GORM配置慢查询记录
func NewDatabase(config *DatabaseConfig, logger *zap.Logger) (*gorm.DB, error) {
    gormLogger := gormlogger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        gormlogger.Config{
            SlowThreshold:             100 * time.Millisecond, // 慢查询阈值
            LogLevel:                  gormlogger.Warn,
            IgnoreRecordNotFoundError: true,
            Colorful:                  true,
        },
    )
    
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: gormLogger,
    })
    
    return db, err
}
```

### 数据库连接池监控
```go
func (app *App) DatabaseStats() gin.HandlerFunc {
    return func(c *gin.Context) {
        sqlDB, err := app.DB.DB()
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        
        stats := sqlDB.Stats()
        c.JSON(200, gin.H{
            "open_connections":     stats.OpenConnections,
            "in_use":              stats.InUse,
            "idle":                stats.Idle,
            "wait_count":          stats.WaitCount,
            "wait_duration":       stats.WaitDuration,
            "max_idle_closed":     stats.MaxIdleClosed,
            "max_lifetime_closed": stats.MaxLifetimeClosed,
        })
    }
}
```

## 📚 相关文档

- [Web框架指南](go-web-framework.md) - Gin和Wire使用
- [可观测性指南](go-observability.md) - 日志和监控
- [权限系统设计](../business/architecture/permission-system.md) - 数据库权限设计
- [数据库模型设计](../business/database/schema-design.md) - 业务数据模型

## 🔗 外部资源

- [GORM官方文档](https://gorm.io/docs/)
- [Redis官方文档](https://redis.io/documentation)
- [MySQL性能优化指南](https://dev.mysql.com/doc/refman/8.0/en/optimization.html)