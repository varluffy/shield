# Go 可观测性与配置指南

UltraFit项目的可观测性技术栈整合指南，包括日志、追踪、监控和配置管理。

## 🎯 可观测性核心原则

### 日志追踪一体化
- 在所有日志中**自动注入TraceID和SpanID**，实现日志与追踪的关联
- 使用**结构化日志**，便于日志聚合和查询
- 实现**统一的日志格式**，包含请求上下文信息
- 通过**中间件自动注入**追踪信息到请求上下文

## 🔧 Zap 日志系统

### Logger配置
```go
package logger

import (
    "context"
    "go.opentelemetry.io/otel/trace"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

type Logger struct {
    *zap.Logger
}

func NewLogger(env string) (*Logger, error) {
    var config zap.Config
    
    if env == "production" {
        config = zap.NewProductionConfig()
        config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
    } else {
        config = zap.NewDevelopmentConfig()
        config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
        config.Development = true
        config.DisableCaller = false
        config.DisableStacktrace = false
    }
    
    // 自定义时间格式
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    
    // 添加调用者信息
    config.EncoderConfig.CallerKey = "caller"
    config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
    
    zapLogger, err := config.Build()
    if err != nil {
        return nil, err
    }
    
    return &Logger{Logger: zapLogger}, nil
}
```

### 上下文日志方法
```go
// 从上下文提取追踪信息的日志方法
func (l *Logger) InfoWithTrace(ctx context.Context, msg string, fields ...zap.Field) {
    fields = append(fields, l.extractTraceFields(ctx)...)
    l.Info(msg, fields...)
}

func (l *Logger) ErrorWithTrace(ctx context.Context, msg string, fields ...zap.Field) {
    fields = append(fields, l.extractTraceFields(ctx)...)
    l.Error(msg, fields...)
}

func (l *Logger) WarnWithTrace(ctx context.Context, msg string, fields ...zap.Field) {
    fields = append(fields, l.extractTraceFields(ctx)...)
    l.Warn(msg, fields...)
}

func (l *Logger) DebugWithTrace(ctx context.Context, msg string, fields ...zap.Field) {
    fields = append(fields, l.extractTraceFields(ctx)...)
    l.Debug(msg, fields...)
}

// 提取追踪信息
func (l *Logger) extractTraceFields(ctx context.Context) []zap.Field {
    span := trace.SpanFromContext(ctx)
    if !span.IsRecording() {
        return nil
    }
    
    spanContext := span.SpanContext()
    return []zap.Field{
        zap.String("trace_id", spanContext.TraceID().String()),
        zap.String("span_id", spanContext.SpanID().String()),
    }
}
```

### 日志中间件
```go
func EnhancedLoggerMiddleware(logger *Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        method := c.Request.Method
        
        // 生成请求ID
        requestID := generateRequestID()
        c.Set("request_id", requestID)
        
        // 记录请求开始
        logger.InfoWithTrace(c.Request.Context(), "Request started",
            zap.String("method", method),
            zap.String("path", path),
            zap.String("request_id", requestID),
            zap.String("client_ip", c.ClientIP()),
            zap.String("user_agent", c.Request.UserAgent()),
        )
        
        c.Next()
        
        // 记录请求结束
        latency := time.Since(start)
        statusCode := c.Writer.Status()
        
        logger.InfoWithTrace(c.Request.Context(), "Request completed",
            zap.String("method", method),
            zap.String("path", path),
            zap.String("request_id", requestID),
            zap.Int("status_code", statusCode),
            zap.Duration("latency", latency),
            zap.Int("response_size", c.Writer.Size()),
        )
    }
}
```

## 📊 OpenTelemetry 追踪

### 追踪初始化
```go
package tracing

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func InitTracing(serviceName, jaegerEndpoint string) (*trace.TracerProvider, error) {
    // 创建Jaeger导出器
    exp, err := jaeger.New(jaeger.WithCollectorEndpoint(
        jaeger.WithEndpoint(jaegerEndpoint),
    ))
    if err != nil {
        return nil, err
    }
    
    // 创建资源
    res, err := resource.New(context.Background(),
        resource.WithAttributes(
            semconv.ServiceNameKey.String(serviceName),
            semconv.ServiceVersionKey.String("1.0.0"),
        ),
    )
    if err != nil {
        return nil, err
    }
    
    // 创建追踪提供者
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),
        trace.WithResource(res),
        trace.WithSampler(trace.AlwaysSample()),
    )
    
    otel.SetTracerProvider(tp)
    return tp, nil
}
```

### 业务追踪示例
```go
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    // 创建span
    ctx, span := otel.Tracer("user-service").Start(ctx, "CreateUser")
    defer span.End()
    
    // 添加span属性
    span.SetAttributes(
        attribute.String("user.email", req.Email),
        attribute.String("user.name", req.Name),
    )
    
    // 记录开始日志
    s.logger.InfoWithTrace(ctx, "Creating user",
        zap.String("email", req.Email),
        zap.String("name", req.Name),
    )
    
    // 业务逻辑
    user, err := s.userRepo.Create(ctx, &User{
        Email: req.Email,
        Name:  req.Name,
    })
    
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "Failed to create user")
        s.logger.ErrorWithTrace(ctx, "Failed to create user", zap.Error(err))
        return nil, err
    }
    
    // 添加成功属性
    span.SetAttributes(attribute.String("user.id", user.UUID))
    span.SetStatus(codes.Ok, "User created successfully")
    
    s.logger.InfoWithTrace(ctx, "User created successfully",
        zap.String("user_id", user.UUID),
    )
    
    return user, nil
}
```

## ⚙️ Viper 配置管理

### 配置结构设计
```go
package config

import (
    "fmt"
    "strings"
    "time"
    
    "github.com/spf13/viper"
)

// 主配置结构
type Config struct {
    App        AppConfig        `mapstructure:"app"`
    Server     ServerConfig     `mapstructure:"server"`
    Database   DatabaseConfig   `mapstructure:"database"`
    Log        LogConfig        `mapstructure:"log"`
    // 可选配置
    Redis      *RedisConfig     `mapstructure:"redis,omitempty"`
    Jaeger     *JaegerConfig    `mapstructure:"jaeger,omitempty"`
    Auth       *AuthConfig      `mapstructure:"auth,omitempty"`
    HTTPClient *HTTPClientConfig `mapstructure:"http_client,omitempty"`
}

// 应用配置
type AppConfig struct {
    Name        string `mapstructure:"name"`
    Version     string `mapstructure:"version"`
    Environment string `mapstructure:"environment"`
    Debug       bool   `mapstructure:"debug"`
    Language    string `mapstructure:"language"`
}

// 服务器配置
type ServerConfig struct {
    Host         string        `mapstructure:"host"`
    Port         int           `mapstructure:"port"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout"`
    WriteTimeout time.Duration `mapstructure:"write_timeout"`
    IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
    CORS         CORSConfig    `mapstructure:"cors"`
}

// 日志配置
type LogConfig struct {
    Level      string `mapstructure:"level"`
    Format     string `mapstructure:"format"`
    Output     string `mapstructure:"output"`
    MaxSize    int    `mapstructure:"max_size"`
    MaxAge     int    `mapstructure:"max_age"`
    MaxBackups int    `mapstructure:"max_backups"`
    Compress   bool   `mapstructure:"compress"`
}
```

### 配置加载函数
```go
func LoadConfig(configPath string) (*Config, error) {
    viper.SetConfigFile(configPath)
    
    // 设置默认值
    setDefaults()
    
    // 绑定环境变量
    viper.AutomaticEnv()
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    // 读取配置文件
    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    // 验证配置
    if err := validateConfig(&config); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    
    return &config, nil
}

func setDefaults() {
    // 应用默认值
    viper.SetDefault("app.name", "ultrafit")
    viper.SetDefault("app.version", "1.0.0")
    viper.SetDefault("app.environment", "development")
    viper.SetDefault("app.debug", true)
    viper.SetDefault("app.language", "zh")
    
    // 服务器默认值
    viper.SetDefault("server.host", "0.0.0.0")
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("server.read_timeout", "30s")
    viper.SetDefault("server.write_timeout", "30s")
    viper.SetDefault("server.idle_timeout", "120s")
    
    // 日志默认值
    viper.SetDefault("log.level", "info")
    viper.SetDefault("log.format", "json")
    viper.SetDefault("log.output", "stdout")
}

func validateConfig(config *Config) error {
    // 验证必需配置
    if config.App.Name == "" {
        return fmt.Errorf("app.name is required")
    }
    
    if config.Server.Port <= 0 || config.Server.Port > 65535 {
        return fmt.Errorf("server.port must be between 1 and 65535")
    }
    
    if config.Database.Host == "" {
        return fmt.Errorf("database.host is required")
    }
    
    return nil
}
```

### 环境变量绑定
```go
// 支持的环境变量
// ULTRAFIT_APP_NAME=ultrafit
// ULTRAFIT_APP_ENVIRONMENT=production
// ULTRAFIT_SERVER_PORT=8080
// ULTRAFIT_DATABASE_HOST=localhost
// ULTRAFIT_DATABASE_PASSWORD=secret
// ULTRAFIT_REDIS_HOST=localhost
// ULTRAFIT_JAEGER_ENDPOINT=http://localhost:14268/api/traces

func bindEnvVariables() {
    // 核心配置环境变量
    viper.BindEnv("app.environment", "GO_ENV", "ULTRAFIT_APP_ENVIRONMENT")
    viper.BindEnv("database.password", "DB_PASSWORD", "ULTRAFIT_DATABASE_PASSWORD")
    viper.BindEnv("auth.jwt_secret", "JWT_SECRET", "ULTRAFIT_JWT_SECRET")
    
    // 可选配置环境变量
    viper.BindEnv("redis.host", "ULTRAFIT_REDIS_HOST")
    viper.BindEnv("redis.password", "ULTRAFIT_REDIS_PASSWORD")
    viper.BindEnv("jaeger.endpoint", "ULTRAFIT_JAEGER_ENDPOINT")
}
```

## 🔧 HTTP 客户端

### Resty客户端配置
```go
package httpclient

import (
    "context"
    "time"
    
    "github.com/go-resty/resty/v2"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
    "go.uber.org/zap"
)

type HTTPClient struct {
    client *resty.Client
    logger *zap.Logger
}

func NewHTTPClient(config *HTTPClientConfig, logger *zap.Logger) *HTTPClient {
    client := resty.New()
    
    // 基础配置
    client.SetTimeout(config.Timeout)
    client.SetRetryCount(config.RetryCount)
    client.SetRetryWaitTime(config.RetryWaitTime)
    client.SetRetryMaxWaitTime(config.RetryMaxWaitTime)
    
    // 添加追踪
    client.GetClient().Transport = otelhttp.NewTransport(client.GetClient().Transport)
    
    // 添加日志中间件
    client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
        ctx := req.Context()
        logger.InfoWithTrace(ctx, "HTTP request started",
            zap.String("method", req.Method),
            zap.String("url", req.URL),
        )
        return nil
    })
    
    client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
        ctx := resp.Request.Context()
        logger.InfoWithTrace(ctx, "HTTP request completed",
            zap.String("method", resp.Request.Method),
            zap.String("url", resp.Request.URL),
            zap.Int("status_code", resp.StatusCode()),
            zap.Duration("duration", resp.Time()),
        )
        return nil
    })
    
    return &HTTPClient{
        client: client,
        logger: logger,
    }
}

func (h *HTTPClient) Get(ctx context.Context, url string) (*resty.Response, error) {
    return h.client.R().SetContext(ctx).Get(url)
}

func (h *HTTPClient) Post(ctx context.Context, url string, body interface{}) (*resty.Response, error) {
    return h.client.R().SetContext(ctx).SetBody(body).Post(url)
}
```

## 📈 监控指标

### 业务指标收集
```go
// 在Service层添加监控指标
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    start := time.Now()
    
    // 业务逻辑
    user, err := s.createUserLogic(ctx, req)
    
    // 记录指标
    duration := time.Since(start)
    
    if err != nil {
        s.logger.ErrorWithTrace(ctx, "User creation failed",
            zap.Error(err),
            zap.Duration("duration", duration),
            zap.String("operation", "create_user"),
        )
        return nil, err
    }
    
    s.logger.InfoWithTrace(ctx, "User creation succeeded",
        zap.Duration("duration", duration),
        zap.String("operation", "create_user"),
        zap.String("user_id", user.UUID),
    )
    
    return user, nil
}
```

### 健康检查端点
```go
func (app *App) HealthCheck() gin.HandlerFunc {
    return func(c *gin.Context) {
        health := gin.H{
            "status":    "ok",
            "timestamp": time.Now().UTC(),
            "version":   app.Config.App.Version,
            "environment": app.Config.App.Environment,
        }
        
        // 检查数据库连接
        if sqlDB, err := app.DB.DB(); err == nil {
            if err := sqlDB.Ping(); err == nil {
                health["database"] = "ok"
            } else {
                health["database"] = "error"
                health["status"] = "degraded"
            }
        }
        
        // 检查Redis连接
        if app.Redis != nil {
            if err := app.Redis.Ping(c.Request.Context()).Err(); err == nil {
                health["redis"] = "ok"
            } else {
                health["redis"] = "error"
                health["status"] = "degraded"
            }
        }
        
        status := http.StatusOK
        if health["status"] == "degraded" {
            status = http.StatusServiceUnavailable
        }
        
        c.JSON(status, health)
    }
}
```

## 📝 配置文件示例

### 开发环境配置
```yaml
# configs/config.dev.yaml
app:
  name: "ultrafit"
  version: "1.0.0"
  environment: "development"
  debug: true
  language: "zh"

server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"
  cors:
    allow_origins: ["*"]
    allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers: ["*"]

database:
  host: "localhost"
  port: 3306
  username: "root"
  password: ""
  database: "ultrafit_dev"
  charset: "utf8mb4"
  max_open_conns: 100
  max_idle_conns: 10

log:
  level: "debug"
  format: "console"
  output: "stdout"

# 可选配置
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  key_prefix: "ultrafit:dev:"

jaeger:
  enabled: true
  endpoint: "http://localhost:14268/api/traces"
```

### 生产环境配置
```yaml
# configs/config.prod.yaml
app:
  name: "ultrafit"
  version: "1.0.0"
  environment: "production"
  debug: false
  language: "zh"

server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"

database:
  host: "${DB_HOST}"
  port: 3306
  username: "${DB_USERNAME}"
  password: "${DB_PASSWORD}"
  database: "${DB_NAME}"
  charset: "utf8mb4"
  max_open_conns: 100
  max_idle_conns: 10

log:
  level: "info"
  format: "json"
  output: "stdout"

redis:
  host: "${REDIS_HOST}"
  port: 6379
  password: "${REDIS_PASSWORD}"
  db: 0
  key_prefix: "ultrafit:prod:"

jaeger:
  enabled: true
  endpoint: "${JAEGER_ENDPOINT}"
```

## 📚 相关文档

- [Web框架指南](go-web-framework.md) - Gin和Wire使用
- [数据库指南](go-database-guide.md) - GORM和Redis使用
- [开发规则约束](../DEVELOPMENT_RULES.md) - 日志规范
- [开发环境配置](../DEVELOPMENT_SETUP.md) - 环境配置

## 🔗 外部资源

- [Zap日志库](https://github.com/uber-go/zap)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Viper配置管理](https://github.com/spf13/viper)
- [Resty HTTP客户端](https://github.com/go-resty/resty)