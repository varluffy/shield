package middleware

import (
	"github.com/google/wire"
	"github.com/varluffy/shield/internal/services"
	"github.com/varluffy/shield/pkg/auth"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/response"
)

// ProviderSet 中间件层的依赖注入Provider集合
var ProviderSet = wire.NewSet(
	// 认证中间件
	NewAuthMiddleware,

	// 权限中间件
	NewPermissionMiddleware,

	// 黑名单认证中间件
	NewBlacklistAuthMiddlewareProvider,
	NewBlacklistLogMiddlewareProvider,
)

// NewAuthMiddleware 创建JWT认证中间件
func NewAuthMiddleware(
	jwtService auth.JWTService,
	permissionService services.PermissionService,
	logger *logger.Logger,
) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:        jwtService,
		permissionService: permissionService,
		logger:            logger,
		responseWriter:    response.NewResponseWriter(logger),
	}
}

// NewBlacklistAuthMiddlewareProvider 创建黑名单鉴权中间件
func NewBlacklistAuthMiddlewareProvider(
	authService services.BlacklistAuthService,
	logger *logger.Logger,
) *BlacklistAuthMiddleware {
	return &BlacklistAuthMiddleware{
		authService:    authService,
		logger:         logger,
		responseWriter: response.NewResponseWriter(logger),
	}
}

// NewBlacklistLogMiddlewareProvider 创建黑名单日志中间件
func NewBlacklistLogMiddlewareProvider(
	authService services.BlacklistAuthService,
	logger *logger.Logger,
) *BlacklistLogMiddleware {
	return &BlacklistLogMiddleware{
		authService: authService,
		logger:      logger,
		sampleRate:  0.01, // 1%采样率
	}
}
