package repositories

import (
	"github.com/google/wire"
)

// ProviderSet Repository层的依赖注入Provider集合
// 包含所有Repository层的构造函数
var ProviderSet = wire.NewSet(
	// User相关Repository
	NewUserRepository,

	// Role相关Repository
	NewRoleRepository,

	// Permission相关Repository
	NewPermissionRepository,
	NewFieldPermissionRepository,

	// Tenant相关Repository
	NewTenantRepository,

	// Permission Audit相关Repository
	NewPermissionAuditRepository,

	// Blacklist相关Repository
	NewBlacklistRepository,
	NewApiCredentialRepository,

	// 这里可以添加其他Repository
	// NewProductRepository,
	// NewOrderRepository,
	// NewCategoryRepository,
)
