// Package routes provides HTTP route configuration and setup.
// It defines all the API endpoints and their corresponding handlers.
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"github.com/varluffy/shield/internal/config"
	"github.com/varluffy/shield/internal/handlers"
	"github.com/varluffy/shield/internal/middleware"
	"github.com/varluffy/shield/pkg/logger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// SetupRoutes 设置路由
func SetupRoutes(
	cfg *config.Config,
	logger *logger.Logger,
	userHandler *handlers.UserHandler,
	captchaHandler *handlers.CaptchaHandler,
	permissionHandler *handlers.PermissionHandler,
	roleHandler *handlers.RoleHandler,
	fieldPermissionHandler *handlers.FieldPermissionHandler,
	blacklistHandler *handlers.BlacklistHandler,
	authMiddleware *middleware.AuthMiddleware,
	blacklistAuthMiddleware *middleware.BlacklistAuthMiddleware,
	blacklistLogMiddleware *middleware.BlacklistLogMiddleware,
) *gin.Engine {
	// 设置Gin模式
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	r := gin.New()

	// 添加全局中间件
	r.Use(middleware.RecoveryMiddleware(logger))
	r.Use(middleware.CORSMiddleware(cfg.Server.CORS))
	r.Use(middleware.EnhancedLoggerMiddleware(logger))

	// 添加OpenTelemetry中间件
	if cfg.Jaeger != nil && cfg.Jaeger.Enabled {
		r.Use(otelgin.Middleware(cfg.App.Name))
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"app":     cfg.App.Name,
			"version": cfg.App.Version,
		})
	})

	// Swagger API文档 (仅在开发环境)
	if cfg.App.Environment == "development" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// API路由组
	api := r.Group("/api/v1")
	{
		// 验证码路由
		if captchaHandler != nil {
			captcha := api.Group("/captcha")
			{
				captcha.GET("/generate", captchaHandler.GenerateCaptcha)
				captcha.POST("/verify", captchaHandler.VerifyCaptcha)
			}
		}

		// 认证路由 (公开接口)
		auth := api.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
			auth.POST("/test-login", userHandler.TestLogin)
			auth.POST("/refresh", userHandler.RefreshToken)
		}

		// 用户管理路由 (需要认证)
		users := api.Group("/users")
		users.Use(authMiddleware.RequireAuth()) // 要求认证
		{
			users.GET("", authMiddleware.RequirePermission("user_list_api"), userHandler.ListUsers)                   // 需要用户列表查看权限
			users.GET("/:uuid", authMiddleware.RequireOwnerOrPermission("uuid", "user_list_api"), userHandler.GetUser) // 用户查看自己信息或拥有用户查看权限
			users.PUT("/:uuid", authMiddleware.RequireOwnerOrPermission("uuid", "user_update_api"), userHandler.UpdateUser) // 用户更新自己信息或拥有用户更新权限
			users.DELETE("/:uuid", authMiddleware.RequirePermission("user_delete_api"), userHandler.DeleteUser)       // 需要用户删除权限
		}

		// 管理员路由 (需要特定权限)
		admin := api.Group("/admin")
		admin.Use(authMiddleware.RequireAuth()) // 要求认证
		{
			admin.POST("/users", authMiddleware.RequirePermission("user_create_api"), userHandler.CreateUser) // 需要用户创建权限
		}

		// 角色管理路由
		roles := api.Group("/roles")
		roles.Use(authMiddleware.RequireAuth()) // 要求认证
		{
			roles.GET("", authMiddleware.RequirePermission("role_list_api"), roleHandler.ListRoles)
			roles.POST("", authMiddleware.RequirePermission("role_create_api"), roleHandler.CreateRole)
			roles.PUT("/:id", authMiddleware.RequirePermission("role_update_api"), roleHandler.UpdateRole)
			roles.DELETE("/:id", authMiddleware.RequirePermission("role_delete_api"), roleHandler.DeleteRole)
			roles.POST("/:id/permissions", authMiddleware.RequirePermission("role_assign_api"), roleHandler.AssignPermissions)
			roles.GET("/:id/permissions", authMiddleware.RequirePermission("role_list_api"), roleHandler.GetRolePermissions)
		}

		// 权限管理路由（统一接口，自动根据用户身份过滤）
		permissions := api.Group("/permissions")
		permissions.Use(authMiddleware.RequireAuth()) // 要求认证
		{
			// 统一的权限查询接口：
			// - 系统管理员：返回所有权限
			// - 租户管理员：只返回租户权限
			permissions.GET("", authMiddleware.RequirePermission("permission_list_api"), permissionHandler.ListPermissions)
			permissions.GET("/tree", authMiddleware.RequirePermission("permission_list_api"), permissionHandler.GetPermissionTree)
		}

		// 字段权限管理路由  
		fieldPermissions := api.Group("/field-permissions")
		fieldPermissions.Use(authMiddleware.RequireAuth()) // 要求认证
		{
			fieldPermissions.GET("/tables/:tableName/fields", authMiddleware.RequirePermission("field_permission_list_api"), fieldPermissionHandler.GetTableFields)
			fieldPermissions.GET("/roles/:roleId/:tableName", authMiddleware.RequirePermission("field_permission_list_api"), fieldPermissionHandler.GetRoleFieldPermissions)
			fieldPermissions.PUT("/roles/:roleId/:tableName", authMiddleware.RequirePermission("field_permission_update_api"), fieldPermissionHandler.UpdateRoleFieldPermissions)
		}

		// 用户权限查询路由
		userPermissions := api.Group("/user")
		userPermissions.Use(authMiddleware.RequireAuth()) // 要求认证
		{
			userPermissions.GET("/permissions", userHandler.GetUserPermissions)       // 获取当前用户权限列表
			userPermissions.GET("/permissions/menu", userHandler.GetUserMenuPermissions) // 获取当前用户菜单权限
			userPermissions.GET("/field-permissions/:tableName", userHandler.GetUserFieldPermissions) // 获取当前用户字段权限
		}

		// 系统管理路由 (只有系统管理员可以访问)
		system := api.Group("/system")
		system.Use(authMiddleware.RequireAuth(), authMiddleware.RequireSystemAdmin()) // 要求认证且为系统管理员
		{
			system.GET("/permissions", permissionHandler.ListSystemPermissions) // 获取系统权限列表
			system.PUT("/permissions/:id", permissionHandler.UpdatePermission)  // 更新权限
		}

		// 黑名单查询API (HMAC鉴权)
		blacklist := api.Group("/blacklist")
		blacklist.Use(blacklistAuthMiddleware.ValidateHMACAuth(), blacklistLogMiddleware.SamplingLogMiddleware())
		{
			blacklist.POST("/check", blacklistHandler.CheckBlacklist) // 检查黑名单
		}

		// 黑名单管理API (JWT鉴权)
		adminBlacklist := api.Group("/admin/blacklist")
		adminBlacklist.Use(authMiddleware.RequireAuth()) // 要求认证
		{
			adminBlacklist.POST("", authMiddleware.RequirePermission("blacklist_create_api"), blacklistHandler.CreateBlacklist)
			adminBlacklist.POST("/import", authMiddleware.RequirePermission("blacklist_import_api"), blacklistHandler.BatchImportBlacklist)
			adminBlacklist.GET("", authMiddleware.RequirePermission("blacklist_list_api"), blacklistHandler.GetBlacklistList)
			adminBlacklist.DELETE("/:id", authMiddleware.RequirePermission("blacklist_delete_api"), blacklistHandler.DeleteBlacklist)
			adminBlacklist.GET("/stats", authMiddleware.RequirePermission("blacklist_stats_api"), blacklistHandler.GetQueryStats)
		}
	}

	return r
}
