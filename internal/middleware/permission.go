// Package middleware provides HTTP middleware for the Gin framework.
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/varluffy/shield/internal/services"
	"github.com/varluffy/shield/pkg/errors"
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

// RequireAnyPermission 要求任意一个权限（OR逻辑）
func (m *AuthMiddleware) RequireAnyPermission(permissionCodes ...string) gin.HandlerFunc {
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

		// 检查是否拥有任意一个权限
		for _, permissionCode := range permissionCodes {
			hasPermission, err := m.permissionService.CheckUserPermission(ctx, userIDStr, tenantIDStr, permissionCode)
			if err != nil {
				m.logger.ErrorWithTrace(ctx, "Failed to check user permission",
					zap.Error(err),
					zap.String("user_id", userIDStr),
					zap.String("permission_code", permissionCode))
				continue // 继续检查下一个权限
			}

			if hasPermission {
				m.logger.DebugWithTrace(ctx, "Permission granted (any)",
					zap.String("user_id", userIDStr),
					zap.String("granted_permission", permissionCode))
				c.Next()
				return
			}
		}

		// 没有任何权限
		m.logger.WarnWithTrace(ctx, "Insufficient permissions (any)",
			zap.String("user_id", userIDStr),
			zap.Strings("required_permissions", permissionCodes))
		m.responseWriter.Error(c, errors.ErrUserPermissionError())
		c.Abort()
	}
}

// RequireAllPermissions 要求所有权限（AND逻辑）
func (m *AuthMiddleware) RequireAllPermissions(permissionCodes ...string) gin.HandlerFunc {
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

		// 检查是否拥有所有权限
		for _, permissionCode := range permissionCodes {
			hasPermission, err := m.permissionService.CheckUserPermission(ctx, userIDStr, tenantIDStr, permissionCode)
			if err != nil {
				m.logger.ErrorWithTrace(ctx, "Failed to check user permission",
					zap.Error(err),
					zap.String("user_id", userIDStr),
					zap.String("permission_code", permissionCode))
				m.responseWriter.Error(c, errors.ErrInternalError("permission check failed"))
				c.Abort()
				return
			}

			if !hasPermission {
				m.logger.WarnWithTrace(ctx, "Insufficient permissions (all)",
					zap.String("user_id", userIDStr),
					zap.String("missing_permission", permissionCode))
				m.responseWriter.Error(c, errors.ErrUserPermissionError())
				c.Abort()
				return
			}
		}

		m.logger.DebugWithTrace(ctx, "All permissions granted",
			zap.String("user_id", userIDStr),
			zap.Strings("permissions", permissionCodes))

		c.Next()
	}
}

// RequirePermissionByHTTPMethod 根据HTTP方法要求不同权限
func (m *AuthMiddleware) RequirePermissionByHTTPMethod(permissionMap map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		permissionCode, exists := permissionMap[method]
		
		if !exists {
			m.logger.WarnWithTrace(c.Request.Context(), "No permission defined for HTTP method",
				zap.String("method", method))
			m.responseWriter.Error(c, errors.ErrForbidden())
			c.Abort()
			return
		}

		// 调用标准权限检查
		permissionMiddleware := m.RequirePermission(permissionCode)
		permissionMiddleware(c)
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

// ValidateAPIPermission API权限验证中间件（支持动态路由匹配）
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

		// 获取请求路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method

		// 构造API权限代码（简单规则：path + method）
		permissionCode := generateAPIPermissionCode(path, method)

		// 检查权限
		hasPermission, err := m.permissionService.CheckUserPermission(ctx, userIDStr, tenantIDStr, permissionCode)
		if err != nil {
			m.logger.ErrorWithTrace(ctx, "Failed to check API permission",
				zap.Error(err),
				zap.String("user_id", userIDStr),
				zap.String("path", path),
				zap.String("method", method),
				zap.String("permission_code", permissionCode))
			// API权限检查失败时不阻止请求，只记录日志
			c.Next()
			return
		}

		if !hasPermission {
			m.logger.WarnWithTrace(ctx, "API permission denied",
				zap.String("user_id", userIDStr),
				zap.String("path", path),
				zap.String("method", method),
				zap.String("permission_code", permissionCode))
			m.responseWriter.Error(c, errors.ErrUserPermissionError())
			c.Abort()
			return
		}

		m.logger.DebugWithTrace(ctx, "API permission granted",
			zap.String("user_id", userIDStr),
			zap.String("path", path),
			zap.String("method", method),
			zap.String("permission_code", permissionCode))

		c.Next()
	}
}

// generateAPIPermissionCode 生成API权限代码
func generateAPIPermissionCode(path, method string) string {
	// 将路径转换为权限代码格式
	// 例如：/api/v1/users/:uuid -> users_detail_api
	// 例如：/api/v1/roles -> roles_list_api (GET) 或 roles_create_api (POST)

	// 移除API前缀
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.TrimPrefix(path, "/api/")

	// 移除尾部斜杠
	path = strings.TrimSuffix(path, "/")

	// 替换路径分隔符为下划线
	resourcePath := strings.ReplaceAll(path, "/", "_")

	// 移除路径参数（:uuid, :id等）
	parts := strings.Split(resourcePath, "_")
	var cleanParts []string
	for _, part := range parts {
		if !strings.HasPrefix(part, ":") && part != "" {
			cleanParts = append(cleanParts, part)
		}
	}

	if len(cleanParts) == 0 {
		return "api_access"
	}

	resource := strings.Join(cleanParts, "_")

	// 根据HTTP方法确定操作类型
	var action string
	switch method {
	case "GET":
		if strings.Contains(path, ":") {
			action = "detail"
		} else {
			action = "list"
		}
	case "POST":
		action = "create"
	case "PUT", "PATCH":
		action = "update"
	case "DELETE":
		action = "delete"
	default:
		action = "access"
	}

	return resource + "_" + action + "_api"
}

// GetFieldPermissions 从上下文获取字段权限
func GetFieldPermissions(c *gin.Context) (map[string]string, bool) {
	permissions, exists := c.Get("field_permissions")
	if !exists {
		return nil, false
	}

	permMap, ok := permissions.(map[string]string)
	return permMap, ok
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