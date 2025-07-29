package services

import (
	"github.com/google/wire"
)

// ProviderSet Service层的依赖注入Provider集合
// 包含所有Service层的构造函数和接口绑定
var ProviderSet = wire.NewSet(
	// User相关Service
	NewUserService,

	// Permission相关Service
	NewPermissionService,

	// Role相关Service
	NewRoleService,

	// FieldPermission相关Service
	NewFieldPermissionService,

	// Cache相关Service
	NewPermissionCacheServiceWithFallback,

	// Audit相关Service
	NewPermissionAuditService,

	// Blacklist相关Service
	NewBlacklistService,
	NewBlacklistAuthService,
	NewApiCredentialService,

	// 这里可以添加其他Service
	// NewProductService,
	// NewOrderService,
	// NewCategoryService,
	// NewAuthService,
)
