// Package logger provides structured logging functionality with tracing support.
// It integrates Zap logger with OpenTelemetry for distributed tracing.
package logger

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogConfig 日志配置（与 internal/config 保持一致）
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

// Logger 日志器包装
type Logger struct {
	*zap.Logger
}

// NewLogger 创建新的日志器（旧版本，保持向后兼容）
func NewLogger(env string) (*Logger, error) {
	// 使用默认配置
	config := &LogConfig{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		MaxSize:    100,
		MaxAge:     7,
		MaxBackups: 10,
		Compress:   true,
	}

	if env != "production" {
		config.Level = "debug"
		config.Format = "console"
	}

	return NewLoggerWithConfig(config)
}

// NewLoggerWithConfig 使用配置创建日志器
func NewLoggerWithConfig(config *LogConfig) (*Logger, error) {
	// 解析日志级别
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 选择编码器
	var encoder zapcore.Encoder
	if config.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 创建输出 WriteSyncer
	writeSyncer := getLogWriter(config)

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建 logger
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1), // 跳过包装层
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return &Logger{Logger: logger}, nil
}

// getLogWriter 获取日志输出器
func getLogWriter(config *LogConfig) zapcore.WriteSyncer {
	switch config.Output {
	case "file":
		// 使用 lumberjack 进行日志轮转
		return zapcore.AddSync(&lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    config.MaxSize,    // 每个日志文件保存的最大尺寸 单位：M
			MaxAge:     config.MaxAge,     // 文件最多保存多少天
			MaxBackups: config.MaxBackups, // 日志文件最多保存多少个备份
			Compress:   config.Compress,   // 是否压缩
		})
	case "both":
		// 同时输出到文件和控制台
		fileWriter := &lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    config.MaxSize,
			MaxAge:     config.MaxAge,
			MaxBackups: config.MaxBackups,
			Compress:   config.Compress,
		}
		return zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			zapcore.AddSync(fileWriter),
		)
	default: // "stdout"
		return zapcore.AddSync(os.Stdout)
	}
}

// WithTraceContext 从上下文中提取追踪信息并记录日志
func (l *Logger) WithTraceContext(ctx context.Context) *zap.Logger {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return l.Logger
	}

	spanContext := span.SpanContext()
	return l.Logger.With(
		zap.String("trace_id", spanContext.TraceID().String()),
		zap.String("span_id", spanContext.SpanID().String()),
	)
}

// InfoWithTrace 带追踪的Info日志
func (l *Logger) InfoWithTrace(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithTraceContext(ctx).Info(msg, fields...)
}

// ErrorWithTrace 带追踪的Error日志
func (l *Logger) ErrorWithTrace(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithTraceContext(ctx).Error(msg, fields...)
}

// WarnWithTrace 带追踪的Warn日志
func (l *Logger) WarnWithTrace(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithTraceContext(ctx).Warn(msg, fields...)
}

// DebugWithTrace 带追踪的Debug日志
func (l *Logger) DebugWithTrace(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithTraceContext(ctx).Debug(msg, fields...)
}

// BusinessLogger 业务操作日志记录器
type BusinessLogger struct {
	logger *Logger
}

// NewBusinessLogger 创建业务日志记录器
func NewBusinessLogger(logger *Logger) *BusinessLogger {
	return &BusinessLogger{logger: logger}
}

// LogUserOperation 记录用户操作日志
func (bl *BusinessLogger) LogUserOperation(ctx context.Context, operation string, userID uint, metadata map[string]any) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.Uint("user_id", userID),
	}

	for key, value := range metadata {
		fields = append(fields, zap.Any(key, value))
	}

	bl.logger.InfoWithTrace(ctx, "User operation completed", fields...)
}

// LogError 记录错误日志
func (bl *BusinessLogger) LogError(ctx context.Context, operation string, err error, metadata map[string]any) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.Error(err),
	}

	for key, value := range metadata {
		fields = append(fields, zap.Any(key, value))
	}

	bl.logger.ErrorWithTrace(ctx, "Operation failed", fields...)
}
