// Package middleware provides HTTP middleware functions.
// This file contains blacklist logging middleware with sampling support.
package middleware

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/varluffy/shield/internal/services"
	"github.com/varluffy/shield/pkg/logger"
	"go.uber.org/zap"
)

// BlacklistLogMiddleware 黑名单日志中间件
type BlacklistLogMiddleware struct {
	authService services.BlacklistAuthService
	logger      *logger.Logger
	sampleRate  float64 // 采样率 (0.0-1.0)
}

// NewBlacklistLogMiddleware 创建黑名单日志中间件
func NewBlacklistLogMiddleware(
	authService services.BlacklistAuthService,
	logger *logger.Logger,
	sampleRate float64,
) *BlacklistLogMiddleware {
	if sampleRate < 0 {
		sampleRate = 0
	}
	if sampleRate > 1 {
		sampleRate = 1
	}

	return &BlacklistLogMiddleware{
		authService: authService,
		logger:      logger,
		sampleRate:  sampleRate,
	}
}

// SamplingLogMiddleware 采样日志中间件
func (m *BlacklistLogMiddleware) SamplingLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 生成请求ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// 记录请求开始（仅在采样时）
		shouldLog := m.shouldLogRequest(c)
		if shouldLog {
			m.logRequestStart(c, requestID, start)
		}

		c.Next()

		// 记录请求完成
		duration := time.Since(start)
		m.logRequestComplete(c, requestID, start, duration, shouldLog)
	}
}

// shouldLogRequest 判断是否应该记录日志
func (m *BlacklistLogMiddleware) shouldLogRequest(c *gin.Context) bool {
	// 错误请求100%记录
	if c.Writer.Status() >= 400 {
		return true
	}

	// 慢请求100%记录（超过50ms）
	if authStart, exists := c.Get("auth_start_time"); exists {
		if startTime, ok := authStart.(time.Time); ok {
			if time.Since(startTime) > 50*time.Millisecond {
				return true
			}
		}
	}

	// 正常请求按采样率记录
	return rand.Float64() < m.sampleRate
}

// logRequestStart 记录请求开始
func (m *BlacklistLogMiddleware) logRequestStart(c *gin.Context, requestID string, start time.Time) {
	ctx := c.Request.Context()

	apiKey, _ := c.Get("api_key")
	tenantID, _ := c.Get("tenant_id")

	m.logger.InfoWithTrace(ctx, "黑名单查询请求开始",
		zap.String("request_id", requestID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("api_key", getStringFromContext(apiKey)),
		zap.Uint64("tenant_id", getUint64FromContext(tenantID)),
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.GetHeader("User-Agent")),
		zap.Time("start_time", start))
}

// logRequestComplete 记录请求完成
func (m *BlacklistLogMiddleware) logRequestComplete(c *gin.Context, requestID string, start time.Time, duration time.Duration, shouldDetailLog bool) {
	ctx := c.Request.Context()

	apiKey := getStringFromContext(c.MustGet("api_key"))
	status := c.Writer.Status()
	responseTime := int(duration.Milliseconds())

	// 确定是否命中黑名单（从响应中判断）
	isHit := false
	if result, exists := c.Get("blacklist_result"); exists {
		if hit, ok := result.(bool); ok {
			isHit = hit
		}
	}

	// 异步记录查询日志到统计系统
	if apiKey != "" {
		phoneMD5 := ""
		if md5, exists := c.Get("phone_md5"); exists {
			phoneMD5 = getStringFromContext(md5)
		}

		m.authService.RecordQueryLog(
			ctx,
			apiKey,
			phoneMD5,
			isHit,
			responseTime,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
			requestID,
		)
	}

	// 根据条件决定是否记录详细日志
	logLevel := zap.InfoLevel
	message := "黑名单查询请求完成"

	if status >= 500 {
		logLevel = zap.ErrorLevel
		message = "黑名单查询请求失败"
		shouldDetailLog = true // 错误请求强制记录详细日志
	} else if status >= 400 {
		logLevel = zap.WarnLevel
		message = "黑名单查询请求错误"
		shouldDetailLog = true // 客户端错误强制记录详细日志
	}

	if shouldDetailLog {
		m.logger.Check(logLevel, message).Write(
			zap.String("request_id", requestID),
			zap.String("api_key", apiKey),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.Int("response_time_ms", responseTime),
			zap.Bool("is_hit", isHit),
			zap.Int("response_size", c.Writer.Size()),
		)
	}

	// 慢查询告警
	if duration > 100*time.Millisecond {
		m.logger.WarnWithTrace(ctx, "慢查询检测",
			zap.String("request_id", requestID),
			zap.String("api_key", apiKey),
			zap.Duration("duration", duration),
			zap.String("path", c.Request.URL.Path))
	}
}

// getStringFromContext 安全地从上下文获取字符串值
func getStringFromContext(value interface{}) string {
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}

// getUint64FromContext 安全地从上下文获取uint64值
func getUint64FromContext(value interface{}) uint64 {
	switch v := value.(type) {
	case uint64:
		return v
	case string:
		if i, err := strconv.ParseUint(v, 10, 64); err == nil {
			return i
		}
	}
	return 0
}
