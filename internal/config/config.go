// Package config provides application configuration management.
// It handles loading, parsing, and validating configuration from various sources.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 主配置结构
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
	// 可选配置 - 只在需要时启用
	Redis      *RedisConfig      `mapstructure:"redis,omitempty"`
	Jaeger     *JaegerConfig     `mapstructure:"jaeger,omitempty"`
	Auth       *AuthConfig       `mapstructure:"auth,omitempty"`
	HTTPClient *HTTPClientConfig `mapstructure:"http_client,omitempty"`
	Captcha    *CaptchaConfig    `mapstructure:"captcha,omitempty"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
	Language    string `mapstructure:"language"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	CORS         CORSConfig    `mapstructure:"cors"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host               string        `mapstructure:"host"`
	Port               int           `mapstructure:"port"`
	User               string        `mapstructure:"user"`
	Password           string        `mapstructure:"password"`
	Name               string        `mapstructure:"name"`
	TimeZone           string        `mapstructure:"timezone"`
	MaxOpenConns       int           `mapstructure:"max_open_conns"`
	MaxIdleConns       int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime    time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime    time.Duration `mapstructure:"conn_max_idle_time"`
	LogLevel           string        `mapstructure:"log_level"`
	SlowQueryThreshold time.Duration `mapstructure:"slow_query_threshold"`
	
	// 迁移控制配置
	EnableAutoMigrate bool   `mapstructure:"enable_auto_migrate"`
	MigrationMode     string `mapstructure:"migration_mode"` // "auto", "manual", "disabled"
}

// RedisConfig Redis配置
type RedisConfig struct {
	// 连接配置
	Addrs    []string `mapstructure:"addrs"`
	Password string   `mapstructure:"password"`
	DB       int      `mapstructure:"db"`

	// 连接池配置
	PoolSize     int `mapstructure:"pool_size"`
	MinIdleConns int `mapstructure:"min_idle_conns"`
	MaxIdleConns int `mapstructure:"max_idle_conns"`

	// 超时配置
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`

	// 业务配置
	KeyPrefix string `mapstructure:"key_prefix"`

	// 可观测性配置
	EnableTracing bool   `mapstructure:"enable_tracing"`
	TracingName   string `mapstructure:"tracing_name"`
}

// JaegerConfig Jaeger配置
type JaegerConfig struct {
	OTLPURL    string  `mapstructure:"otlp_url"`
	SampleRate float64 `mapstructure:"sample_rate"`
	Enabled    bool    `mapstructure:"enabled"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWT            JWTConfig `mapstructure:"jwt"`
	CaptchaMode    string    `mapstructure:"captcha_mode"`    // "strict", "flexible", "disabled"
	DevBypassCode  string    `mapstructure:"dev_bypass_code"` // 开发环境绕过验证码
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret         string        `mapstructure:"secret"`
	ExpiresIn      time.Duration `mapstructure:"expires_in"`
	RefreshExpires time.Duration `mapstructure:"refresh_expires"`
	Issuer         string        `mapstructure:"issuer"`
}

// HTTPClientConfig HTTP客户端配置
type HTTPClientConfig struct {
	// Timeout 请求超时时间 (秒)
	Timeout int `mapstructure:"timeout" default:"30"`

	// RetryCount 重试次数
	RetryCount int `mapstructure:"retry_count" default:"3"`

	// RetryWaitTime 重试等待时间 (秒)
	RetryWaitTime int `mapstructure:"retry_wait_time" default:"1"`

	// RetryMaxWaitTime 最大重试等待时间 (秒)
	RetryMaxWaitTime int `mapstructure:"retry_max_wait_time" default:"10"`

	// EnableTrace 是否启用链路追踪
	EnableTrace bool `mapstructure:"enable_trace" default:"true"`

	// EnableLog 是否启用请求日志
	EnableLog bool `mapstructure:"enable_log" default:"true"`

	// MaxLogBodySize 日志中记录的请求/响应体最大大小 (字节)
	MaxLogBodySize int `mapstructure:"max_log_body_size" default:"10240"` // 10KB

	// UserAgent 用户代理
	UserAgent string `mapstructure:"user_agent" default:"UltraFit-HTTP-Client/1.0"`
}

// CaptchaConfig 验证码配置
type CaptchaConfig struct {
	// Enabled 是否启用验证码
	Enabled bool `mapstructure:"enabled" default:"true"`

	// Type 验证码类型: digit, string, math
	Type string `mapstructure:"type" default:"digit"`

	// Width 图片宽度
	Width int `mapstructure:"width" default:"160"`

	// Height 图片高度
	Height int `mapstructure:"height" default:"60"`

	// Length 验证码长度
	Length int `mapstructure:"length" default:"4"`

	// NoiseCount 噪点数量
	NoiseCount int `mapstructure:"noise_count" default:"5"`

	// Expiration 过期时间
	Expiration time.Duration `mapstructure:"expiration" default:"5m"`
}

// ConfigLoader 配置加载器
type ConfigLoader struct {
	viper *viper.Viper
}

// NewConfigLoader 创建配置加载器
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		viper: viper.New(),
	}
}

// LoadConfig 加载配置
func (c *ConfigLoader) LoadConfig(configPath string) (*Config, error) {
	// 设置配置文件路径和名称
	if configPath != "" {
		c.viper.SetConfigFile(configPath)
	} else {
		// 默认配置文件查找路径
		c.viper.SetConfigName("config")
		c.viper.SetConfigType("yaml")
		c.viper.AddConfigPath(".")
		c.viper.AddConfigPath("./configs")
		c.viper.AddConfigPath("./config")
		c.viper.AddConfigPath("/etc/shield")
	}

	// 设置环境变量
	c.setupEnvironmentVariables()

	// 设置默认值
	c.setDefaults()

	// 读取配置文件
	if err := c.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件未找到，使用默认值和环境变量
			fmt.Println("Config file not found, using defaults and environment variables")
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 解析配置到结构体
	var cfg Config
	if err := c.viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 验证配置
	if err := c.validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// setupEnvironmentVariables 设置环境变量
func (c *ConfigLoader) setupEnvironmentVariables() {
	c.viper.AutomaticEnv()
	c.viper.SetEnvPrefix("ULTRAFIT")
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 绑定核心环境变量
	c.viper.BindEnv("app.environment", "GO_ENV")
	c.viper.BindEnv("database.password", "DB_PASSWORD")
	c.viper.BindEnv("auth.jwt.secret", "JWT_SECRET")
}

// setDefaults 设置默认值
func (c *ConfigLoader) setDefaults() {
	// 应用默认值
	c.viper.SetDefault("app.name", "UltraFit")
	c.viper.SetDefault("app.version", "1.0.0")
	c.viper.SetDefault("app.environment", "development")
	c.viper.SetDefault("app.debug", false)
	c.viper.SetDefault("app.language", "zh")

	// 服务器默认值
	c.viper.SetDefault("server.host", "0.0.0.0")
	c.viper.SetDefault("server.port", 8080)
	c.viper.SetDefault("server.read_timeout", "30s")
	c.viper.SetDefault("server.write_timeout", "30s")
	c.viper.SetDefault("server.idle_timeout", "60s")

	// 数据库默认值
	c.viper.SetDefault("database.host", "localhost")
	c.viper.SetDefault("database.port", 3306)
	c.viper.SetDefault("database.user", "root")
	c.viper.SetDefault("database.name", "shield")
	c.viper.SetDefault("database.timezone", "Asia/Shanghai")
	c.viper.SetDefault("database.max_open_conns", 10)
	c.viper.SetDefault("database.max_idle_conns", 5)
	c.viper.SetDefault("database.slow_query_threshold", "100ms")

	c.viper.SetDefault("database.conn_max_lifetime", "1h")
	c.viper.SetDefault("database.log_level", "info")

	// 日志默认值
	c.viper.SetDefault("log.level", "info")
	c.viper.SetDefault("log.format", "json")
	c.viper.SetDefault("log.output", "stdout")
}

// validateConfig 验证配置
func (c *ConfigLoader) validateConfig(cfg *Config) error {
	// 验证数据库密码
	if cfg.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}

	// 验证端口范围
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}

	// 验证JWT Secret（如果启用了Auth）
	if cfg.Auth != nil && cfg.Auth.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required when auth is enabled")
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
