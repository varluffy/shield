// Package middleware provides HTTP middleware functions.
// This file contains blacklist authentication middleware for HMAC signature validation.
package middleware

import (
	"bytes"
	"io"
	"net"
	"strings"
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

		// IP白名单检查
		clientIP := m.getClientIP(c)
		if !m.isIPAllowed(clientIP, credential.IPWhitelist) {
			m.logger.WarnWithTrace(ctx, "IP地址不在白名单中",
				zap.String("api_key", apiKey),
				zap.String("client_ip", clientIP),
				zap.String("ip_whitelist", credential.IPWhitelist))
			m.responseWriter.Error(c, errors.ErrForbidden())
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

// getClientIP 获取客户端真实IP地址
func (m *BlacklistAuthMiddleware) getClientIP(c *gin.Context) string {
	// 优先从 X-Real-IP 获取
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}

	// 从 X-Forwarded-For 获取（取第一个IP）
	if ips := c.GetHeader("X-Forwarded-For"); ips != "" {
		parts := strings.Split(ips, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	// 从 RemoteAddr 获取
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

// isIPAllowed 检查IP是否在白名单中
func (m *BlacklistAuthMiddleware) isIPAllowed(clientIP, whitelist string) bool {
	// 如果白名单为空，表示不限制IP
	if whitelist == "" {
		return true
	}

	// 解析客户端IP
	clientIPAddr := net.ParseIP(clientIP)
	if clientIPAddr == nil {
		return false
	}

	// 分割白名单中的IP/CIDR
	allowedIPs := strings.Split(whitelist, ",")
	for _, allowedIP := range allowedIPs {
		allowedIP = strings.TrimSpace(allowedIP)
		if allowedIP == "" {
			continue
		}

		// 检查是否为CIDR格式
		if strings.Contains(allowedIP, "/") {
			_, ipNet, err := net.ParseCIDR(allowedIP)
			if err != nil {
				m.logger.Warn("无效的CIDR格式", zap.String("cidr", allowedIP))
				continue
			}
			if ipNet.Contains(clientIPAddr) {
				return true
			}
		} else {
			// 单个IP地址
			allowedIPAddr := net.ParseIP(allowedIP)
			if allowedIPAddr != nil && allowedIPAddr.Equal(clientIPAddr) {
				return true
			}
		}
	}

	return false
}
