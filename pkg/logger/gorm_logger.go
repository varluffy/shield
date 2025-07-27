// Package logger provides GORM logger implementation with Zap and TraceID support.
package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// GormLogger GORM日志器，基于Zap实现
type GormLogger struct {
	ZapLogger                 *zap.Logger
	LogLevel                  logger.LogLevel
	SlowThreshold             time.Duration
	SkipCallerLookup          bool
	IgnoreRecordNotFoundError bool
}

// NewGormLogger 创建新的GORM日志器
func NewGormLogger(zapLogger *zap.Logger) *GormLogger {
	return &GormLogger{
		ZapLogger:                 zapLogger,
		LogLevel:                  logger.Info,
		SlowThreshold:             200 * time.Millisecond,
		SkipCallerLookup:          false,
		IgnoreRecordNotFoundError: true,
	}
}

// LogMode 设置日志模式
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info 信息日志
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.logWithTrace(ctx, zap.InfoLevel, msg, data...)
	}
}

// Warn 警告日志
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.logWithTrace(ctx, zap.WarnLevel, msg, data...)
	}
}

// Error 错误日志
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.logWithTrace(ctx, zap.ErrorLevel, msg, data...)
	}
}

// Trace SQL执行日志
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 构建基础字段
	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
	}

	// 添加调用位置信息
	if !l.SkipCallerLookup {
		fields = append(fields, zap.String("file", utils.FileWithLineNum()))
	}

	// 添加TraceID和SpanID
	fields = append(fields, l.extractTraceFields(ctx)...)

	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		// SQL执行错误
		fields = append(fields, zap.Error(err))
		l.ZapLogger.Error("SQL execution failed", fields...)

	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		// 慢查询警告
		fields = append(fields,
			zap.Duration("slow_threshold", l.SlowThreshold),
			zap.String("performance", "SLOW QUERY"),
		)
		l.ZapLogger.Warn("Slow SQL query detected", fields...)

	case l.LogLevel == logger.Info:
		// 正常SQL执行日志
		fields = append(fields, zap.String("status", "success"))
		l.ZapLogger.Info("SQL executed", fields...)
	}
}

// logWithTrace 带追踪信息的日志记录
func (l *GormLogger) logWithTrace(ctx context.Context, level zapcore.Level, msg string, data ...interface{}) {
	// 格式化消息
	if len(data) > 0 {
		msg = fmt.Sprintf(msg, data...)
	}

	// 提取追踪字段
	fields := l.extractTraceFields(ctx)

	// 根据日志级别记录
	switch level {
	case zap.InfoLevel:
		l.ZapLogger.Info(msg, fields...)
	case zap.WarnLevel:
		l.ZapLogger.Warn(msg, fields...)
	case zap.ErrorLevel:
		l.ZapLogger.Error(msg, fields...)
	case zap.DebugLevel:
		l.ZapLogger.Debug(msg, fields...)
	}
}

// extractTraceFields 从context中提取追踪字段
func (l *GormLogger) extractTraceFields(ctx context.Context) []zap.Field {
	var fields []zap.Field

	// 提取OpenTelemetry追踪信息
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		spanContext := span.SpanContext()
		fields = append(fields,
			zap.String("trace_id", spanContext.TraceID().String()),
			zap.String("span_id", spanContext.SpanID().String()),
		)
	}

	return fields
}

// GormLoggerConfig GORM日志器配置
type GormLoggerConfig struct {
	LogLevel                  logger.LogLevel
	SlowThreshold             time.Duration
	SkipCallerLookup          bool
	IgnoreRecordNotFoundError bool
}

// NewGormLoggerWithConfig 使用配置创建GORM日志器
func NewGormLoggerWithConfig(zapLogger *zap.Logger, config GormLoggerConfig) *GormLogger {
	return &GormLogger{
		ZapLogger:                 zapLogger,
		LogLevel:                  config.LogLevel,
		SlowThreshold:             config.SlowThreshold,
		SkipCallerLookup:          config.SkipCallerLookup,
		IgnoreRecordNotFoundError: config.IgnoreRecordNotFoundError,
	}
}

// ParseGormLogLevel 解析GORM日志级别
func ParseGormLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Info
	}
}
