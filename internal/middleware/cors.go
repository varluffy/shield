package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/varluffy/shield/internal/config"
)

// CORSMiddleware CORS中间件
func CORSMiddleware(corsConfig config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查Origin是否被允许
		if isAllowedOrigin(origin, corsConfig.AllowOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		// 设置允许的方法
		if len(corsConfig.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowMethods, ", "))
		}

		// 设置允许的头部
		if len(corsConfig.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowHeaders, ", "))
		}

		// 设置是否允许携带凭据
		if corsConfig.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 设置预检请求的缓存时间
		c.Header("Access-Control-Max-Age", "86400") // 24小时

		// 如果是预检请求，直接返回
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// isAllowedOrigin 检查Origin是否被允许
func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
		// 支持通配符匹配（简单实现）
		if strings.Contains(allowed, "*") {
			pattern := strings.Replace(allowed, "*", "", -1)
			if strings.Contains(origin, pattern) {
				return true
			}
		}
	}
	return false
}
