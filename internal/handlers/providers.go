package handlers

import (
	"github.com/google/wire"
	"github.com/varluffy/shield/pkg/response"
)

// ProviderSet Handler层的依赖注入Provider集合
// 包含所有Handler层的构造函数
var ProviderSet = wire.NewSet(
	// 响应处理器
	response.NewResponseWriter,

	// User相关Handler
	NewUserHandler,

	// Captcha相关Handler
	NewCaptchaHandler,

	// 权限管理相关Handler
	NewPermissionHandler,
	NewRoleHandler,
	NewFieldPermissionHandler,

	// 这里可以添加其他Handler
	// NewProductHandler,
	// NewOrderHandler,
	// NewCategoryHandler,
	// NewAuthHandler,
) 