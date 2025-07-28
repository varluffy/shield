// Package middleware provides HTTP middleware functions.
// It includes authentication, authorization, and other request processing middleware.
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/varluffy/shield/internal/services"
	"github.com/varluffy/shield/pkg/auth"
	"github.com/varluffy/shield/pkg/errors"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/response"
	"go.uber.org/zap"
)

// AuthMiddleware JWT认证中间件
type AuthMiddleware struct {
	jwtService        auth.JWTService
	permissionService services.PermissionService
	logger            *logger.Logger
	responseWriter    *response.ResponseWriter
}

// RequireAuth 要求用户认证
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 从Header中获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.WarnWithTrace(ctx, "Missing authorization header")
			m.responseWriter.Error(c, errors.ErrUnauthorized())
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			m.logger.WarnWithTrace(ctx, "Invalid authorization header format")
			m.responseWriter.Error(c, errors.ErrUnauthorized())
			c.Abort()
			return
		}

		// 提取token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			m.logger.WarnWithTrace(ctx, "Empty token")
			m.responseWriter.Error(c, errors.ErrUnauthorized())
			c.Abort()
			return
		}

		// 验证token
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			m.logger.WarnWithTrace(ctx, "Invalid token",
				zap.Error(err))
			m.responseWriter.Error(c, errors.ErrInvalidToken())
			c.Abort()
			return
		}

		// 从Header中获取租户ID（可选，如果没有提供则使用token中的租户ID）
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			tenantID = claims.TenantID
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("tenant_id", tenantID)
		c.Set("jwt_claims", claims)

		m.logger.DebugWithTrace(ctx, "User authenticated",
			zap.String("user_id", claims.UserID),
			zap.String("email", claims.Email),
			zap.String("tenant_id", tenantID))

		c.Next()
	}
}

// ValidateAPIPermission 基于API路径和HTTP方法动态验证权限
func (m *AuthMiddleware) ValidateAPIPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 检查是否已认证
		userID, exists := c.Get("user_id")
		if !exists {
			m.logger.WarnWithTrace(ctx, "User ID not found in context")
			m.responseWriter.Error(c, errors.ErrUnauthorized())
			c.Abort()
			return
		}

		tenantID, exists := c.Get("tenant_id")
		if !exists {
			m.logger.WarnWithTrace(ctx, "Tenant ID not found in context")
			m.responseWriter.Error(c, errors.ErrUnauthorized())
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			m.logger.WarnWithTrace(ctx, "Invalid user ID type")
			m.responseWriter.Error(c, errors.ErrForbidden())
			c.Abort()
			return
		}

		tenantIDStr, ok := tenantID.(string)
		if !ok {
			m.logger.WarnWithTrace(ctx, "Invalid tenant ID type")
			m.responseWriter.Error(c, errors.ErrForbidden())
			c.Abort()
			return
		}

		// 检查是否为系统租户用户（tenant_id = 0）
		// 系统租户拥有所有权限，直接放行
		if tenantIDStr == "0" {
			m.logger.DebugWithTrace(ctx, "System tenant user granted all permissions",
				zap.String("user_id", userIDStr),
				zap.String("tenant_id", tenantIDStr))
			c.Next()
			return
		}

		// 获取请求路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method

		// 标准化路径（移除版本前缀）
		path = strings.TrimPrefix(path, "/api/v1")

		// 检查用户是否有访问该API的权限
		hasPermission, err := m.permissionService.CheckUserAPIPermission(ctx, userIDStr, tenantIDStr, path, method)
		if err != nil {
			m.logger.ErrorWithTrace(ctx, "Failed to check API permission",
				zap.Error(err),
				zap.String("user_id", userIDStr),
				zap.String("tenant_id", tenantIDStr),
				zap.String("path", path),
				zap.String("method", method))
			m.responseWriter.Error(c, errors.ErrInternalError("permission check failed"))
			c.Abort()
			return
		}

		if !hasPermission {
			m.logger.WarnWithTrace(ctx, "Insufficient API permissions",
				zap.String("user_id", userIDStr),
				zap.String("tenant_id", tenantIDStr),
				zap.String("path", path),
				zap.String("method", method))
			m.responseWriter.Error(c, errors.ErrUserPermissionError())
			c.Abort()
			return
		}

		m.logger.DebugWithTrace(ctx, "API permission granted",
			zap.String("user_id", userIDStr),
			zap.String("path", path),
			zap.String("method", method))

		c.Next()
	}
}


// OptionalAuth 可选认证中间件
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 从Header中获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有token，跳过认证但继续处理
			c.Next()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		// 提取token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.Next()
			return
		}

		// 尝试验证token
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			m.logger.DebugWithTrace(ctx, "Optional auth: invalid token",
				zap.Error(err))
			c.Next()
			return
		}

		// 从Header中获取租户ID（可选）
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			tenantID = claims.TenantID
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("tenant_id", tenantID)
		c.Set("jwt_claims", claims)

		m.logger.DebugWithTrace(ctx, "Optional auth: user authenticated",
			zap.String("user_id", claims.UserID),
			zap.String("email", claims.Email),
			zap.String("tenant_id", tenantID))

		c.Next()
	}
}

// GetCurrentUser 从上下文中获取当前用户信息
func GetCurrentUser(c *gin.Context) (userID, email, tenantID string, exists bool) {
	userIDVal, userIDExists := c.Get("user_id")
	emailVal, emailExists := c.Get("user_email")
	tenantIDVal, tenantIDExists := c.Get("tenant_id")

	if !userIDExists || !emailExists || !tenantIDExists {
		return "", "", "", false
	}

	userIDStr, ok1 := userIDVal.(string)
	emailStr, ok2 := emailVal.(string)
	tenantIDStr, ok3 := tenantIDVal.(string)

	if !ok1 || !ok2 || !ok3 {
		return "", "", "", false
	}

	return userIDStr, emailStr, tenantIDStr, true
}

// GetCurrentUserID 从上下文中获取当前用户ID
func GetCurrentUserID(c *gin.Context) (string, bool) {
	userID, _, _, exists := GetCurrentUser(c)
	return userID, exists
}

// GetCurrentTenantID 从上下文中获取当前租户ID
func GetCurrentTenantID(c *gin.Context) (string, bool) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return "", false
	}
	tenantIDStr, ok := tenantID.(string)
	return tenantIDStr, ok
}

