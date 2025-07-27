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