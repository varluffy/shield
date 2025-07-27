// Package test contains integration tests and API examples.
// It provides comprehensive testing scenarios for the application.
package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/varluffy/shield/pkg/logger"
	"go.uber.org/zap"
)

func TestLoggerConfiguration(t *testing.T) {
	// 确保测试日志目录存在
	testLogsDir := "test_logs"
	defer os.RemoveAll(testLogsDir)

	tests := []struct {
		name   string
		config *logger.LogConfig
	}{
		{
			name: "Console Output",
			config: &logger.LogConfig{
				Level:      "debug",
				Format:     "console",
				Output:     "stdout",
				MaxSize:    10,
				MaxAge:     1,
				MaxBackups: 3,
				Compress:   false,
			},
		},
		{
			name: "JSON Format",
			config: &logger.LogConfig{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				MaxSize:    10,
				MaxAge:     1,
				MaxBackups: 3,
				Compress:   false,
			},
		},
		{
			name: "File Output",
			config: &logger.LogConfig{
				Level:      "debug",
				Format:     "json",
				Output:     "file",
				MaxSize:    1, // 1MB for testing
				MaxAge:     1,
				MaxBackups: 3,
				Compress:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 logger
			l, err := logger.NewLoggerWithConfig(tt.config)
			assert.NoError(t, err)
			assert.NotNil(t, l)

			// 测试各种日志级别
			ctx := context.Background()

			l.DebugWithTrace(ctx, "Debug message", zap.String("test", "debug"))
			l.InfoWithTrace(ctx, "Info message", zap.String("test", "info"))
			l.WarnWithTrace(ctx, "Warn message", zap.String("test", "warn"))
			l.ErrorWithTrace(ctx, "Error message", zap.String("test", "error"))

			// 如果是文件输出，检查文件是否存在
			if tt.config.Output == "file" {
				time.Sleep(100 * time.Millisecond) // 等待文件写入
				logPath := filepath.Join("logs", "app.log")
				_, err := os.Stat(logPath)
				// 文件可能存在也可能不存在，取决于日志级别
				if err == nil {
					t.Logf("Log file created: %s", logPath)
				}
			}
		})
	}
}

func TestLoggerLevels(t *testing.T) {
	config := &logger.LogConfig{
		Level:      "warn", // 只记录 warn 及以上级别
		Format:     "json",
		Output:     "stdout",
		MaxSize:    10,
		MaxAge:     1,
		MaxBackups: 3,
		Compress:   false,
	}

	l, err := logger.NewLoggerWithConfig(config)
	assert.NoError(t, err)

	ctx := context.Background()

	// 这些不应该输出（低于 warn 级别）
	l.DebugWithTrace(ctx, "Debug should not appear")
	l.InfoWithTrace(ctx, "Info should not appear")

	// 这些应该输出
	l.WarnWithTrace(ctx, "Warn should appear")
	l.ErrorWithTrace(ctx, "Error should appear")
}

func TestBusinessLogger(t *testing.T) {
	config := &logger.LogConfig{
		Level:      "debug",
		Format:     "json",
		Output:     "stdout",
		MaxSize:    10,
		MaxAge:     1,
		MaxBackups: 3,
		Compress:   false,
	}

	baseLogger, err := logger.NewLoggerWithConfig(config)
	assert.NoError(t, err)

	businessLogger := logger.NewBusinessLogger(baseLogger)
	ctx := context.Background()

	// 测试用户操作日志
	businessLogger.LogUserOperation(ctx, "create_user", 123, map[string]interface{}{
		"email": "test@example.com",
		"name":  "Test User",
	})

	// 测试错误日志
	businessLogger.LogError(ctx, "database_error", assert.AnError, map[string]interface{}{
		"table": "users",
		"query": "INSERT INTO users...",
	})
}

func TestBackwardCompatibility(t *testing.T) {
	// 测试旧的 NewLogger 函数是否仍然工作
	l1, err := logger.NewLogger("development")
	assert.NoError(t, err)
	assert.NotNil(t, l1)

	l2, err := logger.NewLogger("production")
	assert.NoError(t, err)
	assert.NotNil(t, l2)

	ctx := context.Background()
	l1.InfoWithTrace(ctx, "Development logger test")
	l2.InfoWithTrace(ctx, "Production logger test")
}
