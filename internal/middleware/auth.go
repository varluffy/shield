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

// RequirePermission 要求特定权限
func (m *AuthMiddleware) RequirePermission(permissionCode string) gin.HandlerFunc {
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

		// 检查权限
		hasPermission, err := m.permissionService.CheckUserPermission(ctx, userIDStr, tenantIDStr, permissionCode)
		if err != nil {
			m.logger.ErrorWithTrace(ctx, "Failed to check user permission",
				zap.Error(err),
				zap.String("user_id", userIDStr),
				zap.String("tenant_id", tenantIDStr),
				zap.String("permission_code", permissionCode))
			m.responseWriter.Error(c, errors.ErrInternalError("permission check failed"))
			c.Abort()
			return
		}

		if !hasPermission {
			m.logger.WarnWithTrace(ctx, "Insufficient permissions",
				zap.String("user_id", userIDStr),
				zap.String("tenant_id", tenantIDStr),
				zap.String("permission_code", permissionCode))
			m.responseWriter.Error(c, errors.ErrUserPermissionError())
			c.Abort()
			return
		}

		m.logger.DebugWithTrace(ctx, "Permission granted",
			zap.String("user_id", userIDStr),
			zap.String("permission_code", permissionCode))

		c.Next()
	}
}

// RequireRole 要求特定角色
func (m *AuthMiddleware) RequireRole(roleCode string) gin.HandlerFunc {
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

		// 检查角色
		hasRole, err := m.permissionService.HasRole(ctx, userIDStr, tenantIDStr, roleCode)
		if err != nil {
			m.logger.ErrorWithTrace(ctx, "Failed to check user role",
				zap.Error(err),
				zap.String("user_id", userIDStr),
				zap.String("tenant_id", tenantIDStr),
				zap.String("role_code", roleCode))
			m.responseWriter.Error(c, errors.ErrInternalError("role check failed"))
			c.Abort()
			return
		}

		if !hasRole {
			m.logger.WarnWithTrace(ctx, "Insufficient role permissions",
				zap.String("user_id", userIDStr),
				zap.String("tenant_id", tenantIDStr),
				zap.String("role_code", roleCode))
			m.responseWriter.Error(c, errors.ErrUserPermissionError())
			c.Abort()
			return
		}

		m.logger.DebugWithTrace(ctx, "Role permission granted",
			zap.String("user_id", userIDStr),
			zap.String("role_code", roleCode))

		c.Next()
	}
}

// RequireSystemAdmin 要求系统管理员权限
func (m *AuthMiddleware) RequireSystemAdmin() gin.HandlerFunc {
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

		userIDStr, ok := userID.(string)
		if !ok {
			m.logger.WarnWithTrace(ctx, "Invalid user ID type")
			m.responseWriter.Error(c, errors.ErrForbidden())
			c.Abort()
			return
		}

		// 检查是否为系统管理员
		isSystemAdmin, err := m.permissionService.IsSystemAdmin(ctx, userIDStr)
		if err != nil {
			m.logger.ErrorWithTrace(ctx, "Failed to check system admin status",
				zap.Error(err),
				zap.String("user_id", userIDStr))
			m.responseWriter.Error(c, errors.ErrInternalError("system admin check failed"))
			c.Abort()
			return
		}

		if !isSystemAdmin {
			m.logger.WarnWithTrace(ctx, "System admin permission required",
				zap.String("user_id", userIDStr))
			m.responseWriter.Error(c, errors.ErrUserPermissionError())
			c.Abort()
			return
		}

		m.logger.DebugWithTrace(ctx, "System admin permission granted",
			zap.String("user_id", userIDStr))

		c.Next()
	}
}

// RequireTenantAdmin 要求租户管理员权限
func (m *AuthMiddleware) RequireTenantAdmin() gin.HandlerFunc {
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

		// 检查是否为租户管理员
		isTenantAdmin, err := m.permissionService.IsTenantAdmin(ctx, userIDStr, tenantIDStr)
		if err != nil {
			m.logger.ErrorWithTrace(ctx, "Failed to check tenant admin status",
				zap.Error(err),
				zap.String("user_id", userIDStr),
				zap.String("tenant_id", tenantIDStr))
			m.responseWriter.Error(c, errors.ErrInternalError("tenant admin check failed"))
			c.Abort()
			return
		}

		if !isTenantAdmin {
			m.logger.WarnWithTrace(ctx, "Tenant admin permission required",
				zap.String("user_id", userIDStr),
				zap.String("tenant_id", tenantIDStr))
			m.responseWriter.Error(c, errors.ErrUserPermissionError())
			c.Abort()
			return
		}

		m.logger.DebugWithTrace(ctx, "Tenant admin permission granted",
			zap.String("user_id", userIDStr),
			zap.String("tenant_id", tenantIDStr))

		c.Next()
	}
}

// RequireOwnerOrAdmin 要求资源所有者或管理员权限
func (m *AuthMiddleware) RequireOwnerOrAdmin(resourceUserIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 获取当前用户ID
		currentUserID, exists := c.Get("user_id")
		if !exists {
			m.logger.WarnWithTrace(ctx, "Current user ID not found in context")
			m.responseWriter.Error(c, errors.ErrUnauthorized())
			c.Abort()
			return
		}

		currentUserIDStr, ok := currentUserID.(string)
		if !ok {
			m.logger.WarnWithTrace(ctx, "Invalid current user ID type")
			m.responseWriter.Error(c, errors.ErrInternalError("invalid user ID"))
			c.Abort()
			return
		}

		// 获取资源用户ID参数
		resourceUserID := c.Param(resourceUserIDParam)
		if resourceUserID == "" {
			m.logger.WarnWithTrace(ctx, "Resource user ID parameter not found",
				zap.String("param_name", resourceUserIDParam))
			m.responseWriter.Error(c, errors.ErrInvalidRequest())
			c.Abort()
			return
		}

		// 如果是资源所有者，直接允许
		if currentUserIDStr == resourceUserID {
			m.logger.DebugWithTrace(ctx, "Owner permission granted",
				zap.String("user_id", currentUserIDStr))
			c.Next()
			return
		}

		// 检查是否为管理员
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			m.logger.WarnWithTrace(ctx, "Tenant ID not found in context")
			m.responseWriter.Error(c, errors.ErrUnauthorized())
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

		// 检查是否为租户管理员或系统管理员
		isTenantAdmin, err := m.permissionService.IsTenantAdmin(ctx, currentUserIDStr, tenantIDStr)
		if err != nil {
			m.logger.ErrorWithTrace(ctx, "Failed to check admin status",
				zap.Error(err),
				zap.String("user_id", currentUserIDStr))
			m.responseWriter.Error(c, errors.ErrInternalError("admin check failed"))
			c.Abort()
			return
		}

		if !isTenantAdmin {
			m.logger.WarnWithTrace(ctx, "Access denied: not resource owner or admin",
				zap.String("current_user_id", currentUserIDStr),
				zap.String("resource_user_id", resourceUserID))
			m.responseWriter.Error(c, errors.ErrUserPermissionError())
			c.Abort()
			return
		}

		m.logger.DebugWithTrace(ctx, "Admin permission granted",
			zap.String("user_id", currentUserIDStr))

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

// RequireOwnerOrPermission 要求资源所有者或特定权限
func (m *AuthMiddleware) RequireOwnerOrPermission(resourceUserIDParam string, permissionCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 获取当前用户ID
		currentUserID, exists := c.Get("user_id")
		if !exists {
			m.logger.WarnWithTrace(ctx, "Current user ID not found in context")
			m.responseWriter.Error(c, errors.ErrUnauthorized())
			c.Abort()
			return
		}

		currentUserIDStr, ok := currentUserID.(string)
		if !ok {
			m.logger.WarnWithTrace(ctx, "Invalid current user ID type")
			m.responseWriter.Error(c, errors.ErrInternalError("invalid user ID"))
			c.Abort()
			return
		}

		// 获取资源用户ID参数
		resourceUserID := c.Param(resourceUserIDParam)
		if resourceUserID == "" {
			m.logger.WarnWithTrace(ctx, "Resource user ID parameter not found",
				zap.String("param_name", resourceUserIDParam))
			m.responseWriter.Error(c, errors.ErrInvalidRequest())
			c.Abort()
			return
		}

		// 如果是资源所有者，直接允许
		if currentUserIDStr == resourceUserID {
			m.logger.DebugWithTrace(ctx, "Owner permission granted",
				zap.String("user_id", currentUserIDStr))
			c.Next()
			return
		}

		// 检查是否有特定权限
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			m.logger.WarnWithTrace(ctx, "Tenant ID not found in context")
			m.responseWriter.Error(c, errors.ErrUnauthorized())
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

		// 检查权限
		hasPermission, err := m.permissionService.CheckUserPermission(ctx, currentUserIDStr, tenantIDStr, permissionCode)
		if err != nil {
			m.logger.ErrorWithTrace(ctx, "Failed to check user permission",
				zap.Error(err),
				zap.String("user_id", currentUserIDStr),
				zap.String("tenant_id", tenantIDStr),
				zap.String("permission_code", permissionCode))
			m.responseWriter.Error(c, errors.ErrInternalError("permission check failed"))
			c.Abort()
			return
		}

		if !hasPermission {
			m.logger.WarnWithTrace(ctx, "Access denied: not resource owner and insufficient permissions",
				zap.String("current_user_id", currentUserIDStr),
				zap.String("resource_user_id", resourceUserID),
				zap.String("permission_code", permissionCode))
			m.responseWriter.Error(c, errors.ErrUserPermissionError())
			c.Abort()
			return
		}

		m.logger.DebugWithTrace(ctx, "Permission granted",
			zap.String("user_id", currentUserIDStr),
			zap.String("permission_code", permissionCode))

		c.Next()
	}
} 