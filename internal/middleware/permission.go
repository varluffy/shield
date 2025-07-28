// Package middleware provides HTTP middleware for the Gin framework.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/varluffy/shield/internal/services"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/response"
	"go.uber.org/zap"
)

// PermissionMiddleware 权限中间件
type PermissionMiddleware struct {
	fieldPermissionService services.FieldPermissionService
	logger                 *logger.Logger
	responseWriter         *response.ResponseWriter
}

// NewPermissionMiddleware 创建权限中间件
func NewPermissionMiddleware(
	fieldPermissionService services.FieldPermissionService,
	logger *logger.Logger,
) *PermissionMiddleware {
	return &PermissionMiddleware{
		fieldPermissionService: fieldPermissionService,
		logger:                 logger,
		responseWriter:         response.NewResponseWriter(logger),
	}
}

// InjectFieldPermissions 注入字段权限信息到上下文
func (pm *PermissionMiddleware) InjectFieldPermissions(tableName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 获取当前用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next() // 如果没有用户信息，跳过字段权限注入
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.Next()
			return
		}

		// 获取租户ID
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.Next()
			return
		}

		tenantIDStr, ok := tenantID.(string)
		if !ok {
			c.Next()
			return
		}

		// 获取用户字段权限
		permissions, err := pm.fieldPermissionService.GetUserFieldPermissions(ctx, userIDStr, tenantIDStr, tableName)
		if err != nil {
			pm.logger.WarnWithTrace(ctx, "Failed to get user field permissions",
				zap.Error(err),
				zap.String("user_id", userIDStr),
				zap.String("table_name", tableName))
			c.Next()
			return
		}

		// 将字段权限注入到上下文
		c.Set("field_permissions", permissions)
		c.Set("field_permissions_table", tableName)

		pm.logger.DebugWithTrace(ctx, "Field permissions injected",
			zap.String("user_id", userIDStr),
			zap.String("table_name", tableName),
			zap.Int("permission_count", len(permissions)))

		c.Next()
	}
}

// GetFieldPermissions 从上下文获取字段权限
func GetFieldPermissions(c *gin.Context) (map[string]string, bool) {
	permissions, exists := c.Get("field_permissions")
	if !exists {
		return nil, false
	}

	permissionMap, ok := permissions.(map[string]string)
	return permissionMap, ok
}

// HasFieldPermission 检查是否有指定字段的权限
func HasFieldPermission(c *gin.Context, fieldName, requiredPermission string) bool {
	permissions, exists := GetFieldPermissions(c)
	if !exists {
		return true // 如果没有字段权限信息，默认允许
	}

	permission, exists := permissions[fieldName]
	if !exists {
		return true // 如果字段没有权限配置，默认允许
	}

	// 权限级别：default > readonly > hidden
	switch requiredPermission {
	case "default":
		return permission == "default"
	case "readonly":
		return permission == "default" || permission == "readonly"
	case "hidden":
		return true // hidden权限总是可以访问（内部使用）
	default:
		return permission == "default"
	}
}
