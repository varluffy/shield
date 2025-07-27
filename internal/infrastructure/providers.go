// Package infrastructure provides dependency injection providers for infrastructure components.
// It contains Wire provider sets for database, logger, tracer and other infrastructure services.
package infrastructure

import (
	"github.com/google/wire"
	"github.com/varluffy/shield/internal/config"
	"github.com/varluffy/shield/internal/database"
	"github.com/varluffy/shield/pkg/httpclient"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/redis"
	"github.com/varluffy/shield/pkg/tracing"
	"github.com/varluffy/shield/pkg/transaction"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProviderSet 基础设施层的依赖注入Provider集合
// 包含配置、日志、数据库、追踪、事务管理、HTTP客户端等基础组件
var ProviderSet = wire.NewSet(
	ProvideConfig,
	ProvideLogger,
	ProvideZapLogger,
	ProvideTracer,
	ProvideDatabase,
	ProvideRedis,
	// 引入事务管理Provider
	transaction.ProviderSet,
	// 引入HTTP客户端Provider
	httpclient.ProviderSet,
)

// ProvideConfig 提供配置
func ProvideConfig(configPath string) (*config.Config, error) {
	loader := config.NewConfigLoader()
	return loader.LoadConfig(configPath)
}

// ProvideLogger 提供日志器
func ProvideLogger(cfg *config.Config) (*logger.Logger, error) {
	// 将配置转换为 logger.LogConfig
	logConfig := &logger.LogConfig{
		Level:      cfg.Log.Level,
		Format:     cfg.Log.Format,
		Output:     cfg.Log.Output,
		MaxSize:    cfg.Log.MaxSize,
		MaxAge:     cfg.Log.MaxAge,
		MaxBackups: cfg.Log.MaxBackups,
		Compress:   cfg.Log.Compress,
	}

	return logger.NewLoggerWithConfig(logConfig)
}

// ProvideTracer 提供追踪器
func ProvideTracer(cfg *config.Config) (func(), error) {
	if !cfg.Jaeger.Enabled {
		return func() {}, nil
	}

	tracingCfg := tracing.Config{
		ServiceName:    cfg.App.Name,
		ServiceVersion: cfg.App.Version,
		Environment:    cfg.App.Environment,
		OTLPURL:        cfg.Jaeger.OTLPURL,
		SampleRate:     cfg.Jaeger.SampleRate,
	}

	return tracing.InitTracer(tracingCfg)
}

// ProvideDatabase 提供数据库连接，支持OpenTelemetry追踪
func ProvideDatabase(cfg *config.Config, logger *logger.Logger) (*gorm.DB, error) {
	db, err := database.NewMySQLConnection(cfg.Database, logger.Logger)
	if err != nil {
		return nil, err
	}

	// 自动迁移
	if err := database.AutoMigrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

// ProvideRedis 提供Redis客户端
func ProvideRedis(cfg *config.Config, logger *logger.Logger) *redis.Client {
	redisConfig := &redis.Config{
		Addrs:         cfg.Redis.Addrs,
		Password:      cfg.Redis.Password,
		DB:            cfg.Redis.DB,
		PoolSize:      cfg.Redis.PoolSize,
		MinIdleConns:  cfg.Redis.MinIdleConns,
		MaxIdleConns:  cfg.Redis.MaxIdleConns,
		DialTimeout:   cfg.Redis.DialTimeout,
		ReadTimeout:   cfg.Redis.ReadTimeout,
		WriteTimeout:  cfg.Redis.WriteTimeout,
		IdleTimeout:   cfg.Redis.IdleTimeout,
		KeyPrefix:     cfg.Redis.KeyPrefix,
		EnableTracing: cfg.Redis.EnableTracing,
		TracingName:   cfg.Redis.TracingName,
	}

	return redis.NewClient(redisConfig, logger.Logger)
}

// ProvideZapLogger 提供原始的zap.Logger（用于需要*zap.Logger的组件）
func ProvideZapLogger(logger *logger.Logger) *zap.Logger {
	return logger.Logger
}
