// Package services contains business logic and cache implementations.
package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/pkg/logger"
	"go.uber.org/zap"
)

// CacheItem 缓存项
type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
}

// IsExpired 检查是否过期
func (c *CacheItem) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// MemoryPermissionCacheService 内存权限缓存服务实现
type MemoryPermissionCacheService struct {
	cache       sync.Map  // key: string, value: *CacheItem
	logger      *logger.Logger
	cachePrefix string
	cacheTTL    time.Duration
	mu          sync.RWMutex
	stopCh      chan struct{}
}

// NewMemoryPermissionCacheService 创建内存权限缓存服务
func NewMemoryPermissionCacheService(logger *logger.Logger) PermissionCacheService {
	service := &MemoryPermissionCacheService{
		logger:      logger,
		cachePrefix: "shield:permission:",
		cacheTTL:    2 * time.Hour, // 2小时缓存
		stopCh:      make(chan struct{}),
	}

	// 启动清理过期缓存的goroutine
	go service.cleanupExpiredItems()

	return service
}

// GetUserPermissions 从缓存获取用户权限
func (s *MemoryPermissionCacheService) GetUserPermissions(ctx context.Context, userID, tenantID string) ([]models.Permission, error) {
	key := s.getUserPermissionKey(userID, tenantID)

	if item, ok := s.cache.Load(key); ok {
		cacheItem := item.(*CacheItem)
		if !cacheItem.IsExpired() {
			if permissions, ok := cacheItem.Data.([]models.Permission); ok {
				s.logger.DebugWithTrace(ctx, "Memory cache hit for user permissions",
					zap.String("user_id", userID),
					zap.String("tenant_id", tenantID),
					zap.Int("permission_count", len(permissions)))
				return permissions, nil
			}
		} else {
			// 删除过期项
			s.cache.Delete(key)
		}
	}

	s.logger.DebugWithTrace(ctx, "Memory cache miss for user permissions",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID))

	return nil, nil // 缓存未命中
}

// SetUserPermissions 设置用户权限到缓存
func (s *MemoryPermissionCacheService) SetUserPermissions(ctx context.Context, userID, tenantID string, permissions []models.Permission) error {
	key := s.getUserPermissionKey(userID, tenantID)

	// 创建副本以避免外部修改影响缓存
	permissionsCopy := make([]models.Permission, len(permissions))
	copy(permissionsCopy, permissions)

	item := &CacheItem{
		Data:      permissionsCopy,
		ExpiresAt: time.Now().Add(s.cacheTTL),
	}

	s.cache.Store(key, item)

	s.logger.DebugWithTrace(ctx, "User permissions cached in memory",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.Int("permission_count", len(permissions)))

	return nil
}

// InvalidateUserPermissions 清除用户权限缓存
func (s *MemoryPermissionCacheService) InvalidateUserPermissions(ctx context.Context, userID, tenantID string) error {
	key := s.getUserPermissionKey(userID, tenantID)

	s.cache.Delete(key)

	s.logger.DebugWithTrace(ctx, "User permissions cache invalidated from memory",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID))

	return nil
}

// GetUserRoles 从缓存获取用户角色
func (s *MemoryPermissionCacheService) GetUserRoles(ctx context.Context, userID, tenantID string) ([]models.Role, error) {
	key := s.getUserRoleKey(userID, tenantID)

	if item, ok := s.cache.Load(key); ok {
		cacheItem := item.(*CacheItem)
		if !cacheItem.IsExpired() {
			if roles, ok := cacheItem.Data.([]models.Role); ok {
				s.logger.DebugWithTrace(ctx, "Memory cache hit for user roles",
					zap.String("user_id", userID),
					zap.String("tenant_id", tenantID),
					zap.Int("role_count", len(roles)))
				return roles, nil
			}
		} else {
			// 删除过期项
			s.cache.Delete(key)
		}
	}

	s.logger.DebugWithTrace(ctx, "Memory cache miss for user roles",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID))

	return nil, nil // 缓存未命中
}

// SetUserRoles 设置用户角色到缓存
func (s *MemoryPermissionCacheService) SetUserRoles(ctx context.Context, userID, tenantID string, roles []models.Role) error {
	key := s.getUserRoleKey(userID, tenantID)

	// 创建副本以避免外部修改影响缓存
	rolesCopy := make([]models.Role, len(roles))
	copy(rolesCopy, roles)

	item := &CacheItem{
		Data:      rolesCopy,
		ExpiresAt: time.Now().Add(s.cacheTTL),
	}

	s.cache.Store(key, item)

	s.logger.DebugWithTrace(ctx, "User roles cached in memory",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.Int("role_count", len(roles)))

	return nil
}

// InvalidateUserRoles 清除用户角色缓存
func (s *MemoryPermissionCacheService) InvalidateUserRoles(ctx context.Context, userID, tenantID string) error {
	key := s.getUserRoleKey(userID, tenantID)

	s.cache.Delete(key)

	s.logger.DebugWithTrace(ctx, "User roles cache invalidated from memory",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID))

	return nil
}

// InvalidateUser 清除用户的所有缓存
func (s *MemoryPermissionCacheService) InvalidateUser(ctx context.Context, userID, tenantID string) error {
	// 清除权限缓存
	s.InvalidateUserPermissions(ctx, userID, tenantID)
	// 清除角色缓存
	s.InvalidateUserRoles(ctx, userID, tenantID)

	s.logger.InfoWithTrace(ctx, "All user cache invalidated from memory",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID))

	return nil
}

// InvalidateAllUserCaches 清除租户所有用户的缓存
func (s *MemoryPermissionCacheService) InvalidateAllUserCaches(ctx context.Context, tenantID string) error {
	prefix := fmt.Sprintf("%suser:%s:", s.cachePrefix, tenantID)
	count := 0

	s.cache.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if len(keyStr) >= len(prefix) && keyStr[:len(prefix)] == prefix {
				s.cache.Delete(key)
				count++
			}
		}
		return true
	})

	s.logger.InfoWithTrace(ctx, "All tenant user caches invalidated from memory",
		zap.String("tenant_id", tenantID),
		zap.Int("invalidated_count", count))

	return nil
}

// GetCacheStats 获取缓存统计信息
func (s *MemoryPermissionCacheService) GetCacheStats(ctx context.Context) map[string]int {
	stats := map[string]int{
		"total_items":   0,
		"expired_items": 0,
		"valid_items":   0,
	}

	now := time.Now()
	s.cache.Range(func(key, value interface{}) bool {
		stats["total_items"]++
		if item, ok := value.(*CacheItem); ok {
			if now.After(item.ExpiresAt) {
				stats["expired_items"]++
			} else {
				stats["valid_items"]++
			}
		}
		return true
	})

	s.logger.DebugWithTrace(ctx, "Memory cache stats",
		zap.Int("total_items", stats["total_items"]),
		zap.Int("expired_items", stats["expired_items"]),
		zap.Int("valid_items", stats["valid_items"]))

	return stats
}

// Stop 停止缓存服务
func (s *MemoryPermissionCacheService) Stop() {
	close(s.stopCh)
}

// cleanupExpiredItems 清理过期的缓存项
func (s *MemoryPermissionCacheService) cleanupExpiredItems() {
	ticker := time.NewTicker(15 * time.Minute) // 每15分钟清理一次过期项
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.performCleanup()
		case <-s.stopCh:
			return
		}
	}
}

// performCleanup 执行清理操作
func (s *MemoryPermissionCacheService) performCleanup() {
	now := time.Now()
	expiredKeys := make([]interface{}, 0)

	// 收集过期的键
	s.cache.Range(func(key, value interface{}) bool {
		if item, ok := value.(*CacheItem); ok {
			if now.After(item.ExpiresAt) {
				expiredKeys = append(expiredKeys, key)
			}
		}
		return true
	})

	// 删除过期的键
	for _, key := range expiredKeys {
		s.cache.Delete(key)
	}

	if len(expiredKeys) > 0 {
		s.logger.DebugWithTrace(context.Background(), "Cleaned up expired cache items",
			zap.Int("expired_count", len(expiredKeys)))
	}
}

// getUserPermissionKey 生成用户权限缓存键
func (s *MemoryPermissionCacheService) getUserPermissionKey(userID, tenantID string) string {
	return fmt.Sprintf("%suser:%s:%s:permissions", s.cachePrefix, tenantID, userID)
}

// getUserRoleKey 生成用户角色缓存键
func (s *MemoryPermissionCacheService) getUserRoleKey(userID, tenantID string) string {
	return fmt.Sprintf("%suser:%s:%s:roles", s.cachePrefix, tenantID, userID)
}