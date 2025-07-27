package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/varluffy/shield/pkg/logger"
	"go.uber.org/zap"
)

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered interface{}) {
		ctx := c.Request.Context()

		// 记录panic信息
		logger.ErrorWithTrace(ctx, "Panic recovered",
			zap.Any("panic", recovered),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("stack", string(debug.Stack())),
		)

		// 返回500错误
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "An unexpected error occurred",
		})
	})
}
