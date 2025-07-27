// Package middleware provides HTTP middleware functions.
// This file contains blacklist authentication middleware for HMAC signature validation.
package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/varluffy/shield/internal/services"
	"github.com/varluffy/shield/pkg/errors"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/response"
	"go.uber.org/zap"
)

// BlacklistAuthMiddleware 黑名单API鉴权中间件
type BlacklistAuthMiddleware struct {
	authService    services.BlacklistAuthService
	logger         *logger.Logger
	responseWriter *response.ResponseWriter
}

// NewBlacklistAuthMiddleware 创建黑名单鉴权中间件
func NewBlacklistAuthMiddleware(
	authService services.BlacklistAuthService,
	logger *logger.Logger,
) *BlacklistAuthMiddleware {
	return &BlacklistAuthMiddleware{
		authService:    authService,
		logger:         logger,
		responseWriter: response.NewResponseWriter(logger),
	}
}

// ValidateHMACAuth HMAC签名验证中间件
func (m *BlacklistAuthMiddleware) ValidateHMACAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		ctx := c.Request.Context()

		// 提取请求头
		apiKey := c.GetHeader("X-API-Key")
		timestamp := c.GetHeader("X-Timestamp")
		nonce := c.GetHeader("X-Nonce")
		signature := c.GetHeader("X-Signature")

		// 检查必需的请求头
		if apiKey == "" {
			m.logger.WarnWithTrace(ctx, "缺少X-API-Key请求头")
			m.responseWriter.Error(c, errors.ErrUnauthorized())
			c.Abort()
			return
		}

		if timestamp == "" {
			m.logger.WarnWithTrace(ctx, "缺少X-Timestamp请求头")
			m.responseWriter.Error(c, errors.ErrUnauthorized())
			c.Abort()
			return
		}

		if nonce == "" {
			m.logger.WarnWithTrace(ctx, "缺少X-Nonce请求头")
			m.responseWriter.Error(c, errors.ErrUnauthorized())
			c.Abort()
			return
		}

		if signature == "" {
			m.logger.WarnWithTrace(ctx, "缺少X-Signature请求头")
			m.responseWriter.Error(c, errors.ErrUnauthorized())
			c.Abort()
			return
		}

		// 读取请求体（用于签名验证）
		body, err := m.readRequestBody(c)
		if err != nil {
			m.logger.ErrorWithTrace(ctx, "读取请求体失败",
				zap.Error(err))
			m.responseWriter.Error(c, errors.ErrInvalidRequest())
			c.Abort()
			return
		}

		// 验证HMAC签名
		credential, err := m.authService.ValidateHMACSignature(ctx, apiKey, timestamp, nonce, signature, body)
		if err != nil {
			m.logger.WarnWithTrace(ctx, "HMAC签名验证失败",
				zap.String("api_key", apiKey),
				zap.Error(err))
			m.responseWriter.Error(c, errors.ErrUnauthorized())
			c.Abort()
			return
		}

		// 检查速率限制
		err = m.authService.CheckRateLimit(ctx, apiKey)
		if err != nil {
			m.logger.WarnWithTrace(ctx, "请求频率超限",
				zap.String("api_key", apiKey),
				zap.Error(err))
			m.responseWriter.Error(c, errors.ErrRateLimitExceeded())
			c.Abort()
			return
		}

		// 设置上下文信息
		c.Set("api_key", apiKey)
		c.Set("tenant_id", credential.TenantID)
		c.Set("credential", credential)
		c.Set("auth_start_time", start)

		// 异步更新API密钥使用时间
		m.authService.UpdateAPIKeyUsage(ctx, apiKey)

		m.logger.DebugWithTrace(ctx, "HMAC鉴权成功",
			zap.String("api_key", apiKey),
			zap.Uint64("tenant_id", credential.TenantID),
			zap.Duration("auth_duration", time.Since(start)))

		c.Next()
	}
}

// readRequestBody 读取请求体并重新设置
func (m *BlacklistAuthMiddleware) readRequestBody(c *gin.Context) (string, error) {
	if c.Request.Body == nil {
		return "", nil
	}

	// 读取请求体
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", err
	}

	// 重新设置请求体（因为已经被读取）
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return string(bodyBytes), nil
}