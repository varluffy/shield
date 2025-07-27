// Package redis provides Redis client with automatic key prefixing and tracing support.
// It uses go-redis hooks for prefix and OpenTelemetry tracing integration.
package redis

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Client Redis客户端，支持自动前缀和链路追踪
type Client struct {
	redis.UniversalClient
	prefix string
	logger *zap.Logger
	tracer trace.Tracer
}

// Config Redis配置
type Config struct {
	// 连接配置
	Addrs    []string `mapstructure:"addrs" json:"addrs"`
	Password string   `mapstructure:"password" json:"password"`
	DB       int      `mapstructure:"db" json:"db"`

	// 连接池配置
	PoolSize     int `mapstructure:"pool_size" json:"pool_size"`
	MinIdleConns int `mapstructure:"min_idle_conns" json:"min_idle_conns"`
	MaxIdleConns int `mapstructure:"max_idle_conns" json:"max_idle_conns"`

	// 超时配置
	DialTimeout  time.Duration `mapstructure:"dial_timeout" json:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" json:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" json:"idle_timeout"`

	// 业务配置
	KeyPrefix string `mapstructure:"key_prefix" json:"key_prefix"`

	// 可观测性配置
	EnableTracing bool   `mapstructure:"enable_tracing" json:"enable_tracing"`
	TracingName   string `mapstructure:"tracing_name" json:"tracing_name"`
}

// NewClient 创建新的Redis客户端
func NewClient(cfg *Config, logger *zap.Logger) *Client {
	// 设置默认值
	if len(cfg.Addrs) == 0 {
		cfg.Addrs = []string{"localhost:6379"}
	}
	if cfg.PoolSize == 0 {
		cfg.PoolSize = 10
	}
	if cfg.MinIdleConns == 0 {
		cfg.MinIdleConns = 2
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 5 * time.Second
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 3 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 3 * time.Second
	}
	if cfg.TracingName == "" {
		cfg.TracingName = "redis"
	}

	// 创建Redis客户端选项
	opt := &redis.UniversalOptions{
		Addrs:           cfg.Addrs,
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		ConnMaxIdleTime: cfg.IdleTimeout,
	}

	// 创建Redis客户端
	rdb := redis.NewUniversalClient(opt)

	// 创建客户端包装器
	client := &Client{
		UniversalClient: rdb,
		prefix:          cfg.KeyPrefix,
		logger:          logger,
	}

	// 设置tracer
	if cfg.EnableTracing {
		client.tracer = otel.Tracer(cfg.TracingName)
	}

	// 添加Hook
	client.addHooks()

	return client
}

// addHooks 添加Redis Hook
func (c *Client) addHooks() {
	// 添加前缀和追踪Hook
	c.AddHook(&prefixTracingHook{
		prefix: c.prefix,
		logger: c.logger,
		tracer: c.tracer,
	})
}

// prefixTracingHook 前缀和追踪Hook
type prefixTracingHook struct {
	prefix string
	logger *zap.Logger
	tracer trace.Tracer
}

// DialHook 连接钩子
func (h *prefixTracingHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if h.tracer != nil {
			var span trace.Span
			ctx, span = h.tracer.Start(ctx, "redis.dial",
				trace.WithAttributes(
					attribute.String("redis.network", network),
					attribute.String("redis.addr", addr),
				))
			defer span.End()

			conn, err := next(ctx, network, addr)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return conn, err
		}

		return next(ctx, network, addr)
	}
}

// ProcessHook 处理钩子
func (h *prefixTracingHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		// 添加前缀
		h.addPrefixToCmd(cmd)

		// 开始追踪
		if h.tracer != nil {
			var span trace.Span
			ctx, span = h.tracer.Start(ctx, fmt.Sprintf("redis.%s", strings.ToLower(cmd.Name())),
				trace.WithAttributes(
					attribute.String("redis.cmd", cmd.Name()),
					attribute.String("redis.args", cmd.String()),
				))
			defer span.End()

			err := next(ctx, cmd)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return err
		}

		return next(ctx, cmd)
	}
}

// ProcessPipelineHook 管道处理钩子
func (h *prefixTracingHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		// 为所有命令添加前缀
		for _, cmd := range cmds {
			h.addPrefixToCmd(cmd)
		}

		// 开始追踪
		if h.tracer != nil {
			var span trace.Span
			ctx, span = h.tracer.Start(ctx, "redis.pipeline",
				trace.WithAttributes(
					attribute.Int("redis.pipeline.length", len(cmds)),
				))
			defer span.End()

			err := next(ctx, cmds)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return err
		}

		return next(ctx, cmds)
	}
}

// addPrefixToCmd 为命令添加前缀
func (h *prefixTracingHook) addPrefixToCmd(cmd redis.Cmder) {
	if h.prefix == "" {
		return
	}

	args := cmd.Args()
	if len(args) < 2 {
		return
	}

	cmdName := strings.ToLower(args[0].(string))

	// 根据不同的命令类型添加前缀
	switch cmdName {
	case "get", "set", "del", "exists", "expire", "ttl", "type", "getset":
		// 单个key的命令
		if len(args) >= 2 {
			if key, ok := args[1].(string); ok {
				args[1] = h.prefix + key
			}
		}
	case "mget", "mset", "msetnx":
		// 多个key的命令
		h.addPrefixToMultiKeys(args, cmdName)
	case "hget", "hset", "hdel", "hexists", "hgetall", "hkeys", "hvals", "hlen", "hmget", "hmset":
		// Hash命令，第一个参数是key
		if len(args) >= 2 {
			if key, ok := args[1].(string); ok {
				args[1] = h.prefix + key
			}
		}
	case "lpush", "rpush", "lpop", "rpop", "llen", "lrange", "lindex", "lset", "lrem", "ltrim":
		// List命令，第一个参数是key
		if len(args) >= 2 {
			if key, ok := args[1].(string); ok {
				args[1] = h.prefix + key
			}
		}
	case "sadd", "srem", "sismember", "smembers", "scard", "spop", "srandmember":
		// Set命令，第一个参数是key
		if len(args) >= 2 {
			if key, ok := args[1].(string); ok {
				args[1] = h.prefix + key
			}
		}
	case "zadd", "zrem", "zscore", "zrange", "zrevrange", "zcard", "zcount", "zrangebyscore":
		// ZSet命令，第一个参数是key
		if len(args) >= 2 {
			if key, ok := args[1].(string); ok {
				args[1] = h.prefix + key
			}
		}
	case "keys", "scan":
		// 模式匹配命令，需要特殊处理
		h.addPrefixToPattern(args, cmdName)
	}
}

// addPrefixToMultiKeys 为多key命令添加前缀
func (h *prefixTracingHook) addPrefixToMultiKeys(args []interface{}, cmdName string) {
	switch cmdName {
	case "mget":
		// MGET key1 key2 key3...
		for i := 1; i < len(args); i++ {
			if key, ok := args[i].(string); ok {
				args[i] = h.prefix + key
			}
		}
	case "mset", "msetnx":
		// MSET key1 value1 key2 value2...
		for i := 1; i < len(args); i += 2 {
			if key, ok := args[i].(string); ok {
				args[i] = h.prefix + key
			}
		}
	}
}

// addPrefixToPattern 为模式匹配命令添加前缀
func (h *prefixTracingHook) addPrefixToPattern(args []interface{}, cmdName string) {
	switch cmdName {
	case "keys":
		// KEYS pattern
		if len(args) >= 2 {
			if pattern, ok := args[1].(string); ok {
				args[1] = h.prefix + pattern
			}
		}
	case "scan":
		// SCAN cursor [MATCH pattern] [COUNT count]
		for i := 1; i < len(args); i++ {
			if str, ok := args[i].(string); ok && strings.ToLower(str) == "match" {
				if i+1 < len(args) {
					if pattern, ok := args[i+1].(string); ok {
						args[i+1] = h.prefix + pattern
					}
				}
				break
			}
		}
	}
}

// GetPrefix 获取前缀
func (c *Client) GetPrefix() string {
	return c.prefix
}

// Health 健康检查
func (c *Client) Health(ctx context.Context) error {
	return c.Ping(ctx).Err()
}

// Stats 获取统计信息
func (c *Client) Stats() *redis.PoolStats {
	return c.PoolStats()
}

// SetPrefix 设置新的前缀（动态修改）
func (c *Client) SetPrefix(prefix string) {
	c.prefix = prefix
	// 重新添加Hook
	c.addHooks()
}
