//go:build wireinject
// +build wireinject

// Package wire contains dependency injection configuration.
// It uses Google Wire to generate dependency injection code for the application.
package wire

import (
	"github.com/google/wire"
	"github.com/varluffy/shield/internal/config"
	"github.com/varluffy/shield/internal/handlers"
	"github.com/varluffy/shield/internal/infrastructure"
	"github.com/varluffy/shield/internal/middleware"
	"github.com/varluffy/shield/internal/repositories"
	"github.com/varluffy/shield/internal/services"
	"github.com/varluffy/shield/pkg/auth"
	"github.com/varluffy/shield/pkg/captcha"
	"github.com/varluffy/shield/pkg/logger"
	"gorm.io/gorm"
)

// ProviderSet 应用程序的完整Provider集合
// 组合了各个模块的ProviderSet，避免重复定义和冲突
var ProviderSet = wire.NewSet(
	// 基础设施层 - 配置、日志、数据库、追踪等
	infrastructure.ProviderSet,

	// Repository层 - 数据访问层的所有Repository
	repositories.ProviderSet,

	// Service层 - 业务逻辑层的所有Service
	services.ProviderSet,

	// Handler层 - HTTP处理层的所有Handler
	handlers.ProviderSet,

	// 中间件层 - 认证、日志等中间件
	middleware.ProviderSet,

	// 认证层 - JWT服务等认证相关组件
	auth.AuthProviderSet,

	// 验证码层 - 验证码相关的所有组件
	captcha.ProviderSet,
)

// ServiceSet 服务层依赖注入集合
var ServiceSet = wire.NewSet(
	infrastructure.ProviderSet,
	repositories.ProviderSet,
	services.ProviderSet,
	handlers.ProviderSet,
)

// InitializeApp 初始化应用
func InitializeApp(configPath string) (*App, func(), error) {
	wire.Build(
		ProviderSet,
		NewApp,
	)
	return &App{}, nil, nil
}

// App 应用结构
type App struct {
	Config                  *config.Config
	Logger                  *logger.Logger
	DB                      *gorm.DB
	UserHandler             *handlers.UserHandler
	CaptchaHandler          *handlers.CaptchaHandler
	PermissionHandler       *handlers.PermissionHandler
	RoleHandler             *handlers.RoleHandler
	FieldPermissionHandler  *handlers.FieldPermissionHandler
	BlacklistHandler        *handlers.BlacklistHandler
	ApiCredentialHandler    *handlers.ApiCredentialHandler
	AuthMiddleware          *middleware.AuthMiddleware
	PermissionMiddleware    *middleware.PermissionMiddleware
	BlacklistAuthMiddleware *middleware.BlacklistAuthMiddleware
	BlacklistLogMiddleware  *middleware.BlacklistLogMiddleware
}

// NewApp 创建应用实例
func NewApp(
	cfg *config.Config,
	logger *logger.Logger,
	db *gorm.DB,
	userHandler *handlers.UserHandler,
	captchaHandler *handlers.CaptchaHandler,
	permissionHandler *handlers.PermissionHandler,
	roleHandler *handlers.RoleHandler,
	fieldPermissionHandler *handlers.FieldPermissionHandler,
	blacklistHandler *handlers.BlacklistHandler,
	apiCredentialHandler *handlers.ApiCredentialHandler,
	authMiddleware *middleware.AuthMiddleware,
	permissionMiddleware *middleware.PermissionMiddleware,
	blacklistAuthMiddleware *middleware.BlacklistAuthMiddleware,
	blacklistLogMiddleware *middleware.BlacklistLogMiddleware,
) *App {
	return &App{
		Config:                  cfg,
		Logger:                  logger,
		DB:                      db,
		UserHandler:             userHandler,
		CaptchaHandler:          captchaHandler,
		PermissionHandler:       permissionHandler,
		RoleHandler:             roleHandler,
		FieldPermissionHandler:  fieldPermissionHandler,
		BlacklistHandler:        blacklistHandler,
		ApiCredentialHandler:    apiCredentialHandler,
		AuthMiddleware:          authMiddleware,
		PermissionMiddleware:    permissionMiddleware,
		BlacklistAuthMiddleware: blacklistAuthMiddleware,
		BlacklistLogMiddleware:  blacklistLogMiddleware,
	}
}
