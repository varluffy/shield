// Package services contains business logic and cache implementations.
package services

import (
	"github.com/varluffy/shield/internal/config"
	"github.com/varluffy/shield/pkg/logger"
	redisClient "github.com/varluffy/shield/pkg/redis"
)

// NewPermissionCacheServiceWithFallback 创建权限缓存服务（带回退机制）
// 如果Redis可用，使用Redis缓存；否则使用内存缓存
func NewPermissionCacheServiceWithFallback(
	cfg *config.Config,
	redisClient *redisClient.Client,
	logger *logger.Logger,
) PermissionCacheService {
	// 如果Redis配置存在且可用，使用Redis缓存
	if cfg.Redis != nil && len(cfg.Redis.Addrs) > 0 && redisClient != nil {
		logger.Info("Using Redis permission cache")
		return NewPermissionCacheService(redisClient, logger)
	}

	// 否则使用内存缓存
	logger.Info("Using memory permission cache (Redis not available or not configured)")
	return NewMemoryPermissionCacheService(logger)
}

// NewPermissionCacheServiceForce 强制创建指定类型的权限缓存服务
func NewPermissionCacheServiceForce(
	cacheType string,
	redisClient *redisClient.Client,
	logger *logger.Logger,
) PermissionCacheService {
	switch cacheType {
	case "redis":
		if redisClient == nil {
			logger.Warn("Redis client is nil, falling back to memory cache")
			return NewMemoryPermissionCacheService(logger)
		}
		logger.Info("Using Redis permission cache (forced)")
		return NewPermissionCacheService(redisClient, logger)
	case "memory":
		logger.Info("Using memory permission cache (forced)")
		return NewMemoryPermissionCacheService(logger)
	default:
		logger.Warn("Unknown cache type, using memory cache")
		return NewMemoryPermissionCacheService(logger)
	}
}