// Package services contains business logic and cache implementations.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/pkg/logger"
	redisClient "github.com/varluffy/shield/pkg/redis"
	"go.uber.org/zap"
)

// PermissionCacheService 权限缓存服务接口
type PermissionCacheService interface {
	// GetUserPermissions 从缓存获取用户权限
	GetUserPermissions(ctx context.Context, userID, tenantID string) ([]models.Permission, error)
	// SetUserPermissions 设置用户权限到缓存
	SetUserPermissions(ctx context.Context, userID, tenantID string, permissions []models.Permission) error
	// InvalidateUserPermissions 清除用户权限缓存
	InvalidateUserPermissions(ctx context.Context, userID, tenantID string) error
	// GetUserRoles 从缓存获取用户角色
	GetUserRoles(ctx context.Context, userID, tenantID string) ([]models.Role, error)
	// SetUserRoles 设置用户角色到缓存
	SetUserRoles(ctx context.Context, userID, tenantID string, roles []models.Role) error
	// InvalidateUserRoles 清除用户角色缓存
	InvalidateUserRoles(ctx context.Context, userID, tenantID string) error
}

// permissionCacheService 权限缓存服务实现
type permissionCacheService struct {
	redisClient *redisClient.Client
	logger      *logger.Logger
	cachePrefix string
	cacheTTL    time.Duration
}

// NewPermissionCacheService 创建权限缓存服务
func NewPermissionCacheService(
	redisClient *redisClient.Client,
	logger *logger.Logger,
) PermissionCacheService {
	return &permissionCacheService{
		redisClient: redisClient,
		logger:      logger,
		cachePrefix: "ultrafit:permission:",
		cacheTTL:    30 * time.Minute, // 30分钟缓存
	}
}

// GetUserPermissions 从缓存获取用户权限
func (s *permissionCacheService) GetUserPermissions(ctx context.Context, userID, tenantID string) ([]models.Permission, error) {
	if s.redisClient == nil {
		return nil, nil // Redis不可用时直接返回nil，让调用方从数据库获取
	}

	key := s.getUserPermissionKey(userID, tenantID)
	
	data, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			s.logger.DebugWithTrace(ctx, "User permissions cache miss",
				zap.String("user_id", userID),
				zap.String("tenant_id", tenantID))
			return nil, nil // 缓存未命中
		}
		s.logger.ErrorWithTrace(ctx, "Failed to get user permissions from cache",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, nil // 发生错误时返回nil，让调用方从数据库获取
	}

	var permissions []models.Permission
	if err := json.Unmarshal([]byte(data), &permissions); err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to unmarshal permissions from cache",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, nil // 反序列化失败时返回nil
	}

	s.logger.DebugWithTrace(ctx, "User permissions cache hit",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.Int("permission_count", len(permissions)))

	return permissions, nil
}

// SetUserPermissions 设置用户权限到缓存
func (s *permissionCacheService) SetUserPermissions(ctx context.Context, userID, tenantID string, permissions []models.Permission) error {
	if s.redisClient == nil {
		return nil // Redis不可用时静默失败
	}

	key := s.getUserPermissionKey(userID, tenantID)
	
	data, err := json.Marshal(permissions)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to marshal permissions for cache",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil // 序列化失败时静默失败
	}

	err = s.redisClient.Set(ctx, key, data, s.cacheTTL).Err()
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to set user permissions to cache",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil // 缓存失败时静默失败
	}

	s.logger.DebugWithTrace(ctx, "User permissions cached",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.Int("permission_count", len(permissions)))

	return nil
}

// InvalidateUserPermissions 清除用户权限缓存
func (s *permissionCacheService) InvalidateUserPermissions(ctx context.Context, userID, tenantID string) error {
	if s.redisClient == nil {
		return nil // Redis不可用时静默失败
	}

	key := s.getUserPermissionKey(userID, tenantID)
	
	err := s.redisClient.Del(ctx, key).Err()
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to invalidate user permissions cache",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil // 删除失败时静默失败
	}

	s.logger.DebugWithTrace(ctx, "User permissions cache invalidated",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID))

	return nil
}

// GetUserRoles 从缓存获取用户角色
func (s *permissionCacheService) GetUserRoles(ctx context.Context, userID, tenantID string) ([]models.Role, error) {
	if s.redisClient == nil {
		return nil, nil // Redis不可用时直接返回nil
	}

	key := s.getUserRoleKey(userID, tenantID)
	
	data, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			s.logger.DebugWithTrace(ctx, "User roles cache miss",
				zap.String("user_id", userID),
				zap.String("tenant_id", tenantID))
			return nil, nil // 缓存未命中
		}
		s.logger.ErrorWithTrace(ctx, "Failed to get user roles from cache",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, nil // 发生错误时返回nil
	}

	var roles []models.Role
	if err := json.Unmarshal([]byte(data), &roles); err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to unmarshal roles from cache",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, nil // 反序列化失败时返回nil
	}

	s.logger.DebugWithTrace(ctx, "User roles cache hit",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.Int("role_count", len(roles)))

	return roles, nil
}

// SetUserRoles 设置用户角色到缓存
func (s *permissionCacheService) SetUserRoles(ctx context.Context, userID, tenantID string, roles []models.Role) error {
	if s.redisClient == nil {
		return nil // Redis不可用时静默失败
	}

	key := s.getUserRoleKey(userID, tenantID)
	
	data, err := json.Marshal(roles)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to marshal roles for cache",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil // 序列化失败时静默失败
	}

	err = s.redisClient.Set(ctx, key, data, s.cacheTTL).Err()
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to set user roles to cache",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil // 缓存失败时静默失败
	}

	s.logger.DebugWithTrace(ctx, "User roles cached",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.Int("role_count", len(roles)))

	return nil
}

// InvalidateUserRoles 清除用户角色缓存
func (s *permissionCacheService) InvalidateUserRoles(ctx context.Context, userID, tenantID string) error {
	if s.redisClient == nil {
		return nil // Redis不可用时静默失败
	}

	key := s.getUserRoleKey(userID, tenantID)
	
	err := s.redisClient.Del(ctx, key).Err()
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to invalidate user roles cache",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil // 删除失败时静默失败
	}

	s.logger.DebugWithTrace(ctx, "User roles cache invalidated",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID))

	return nil
}

// getUserPermissionKey 生成用户权限缓存键
func (s *permissionCacheService) getUserPermissionKey(userID, tenantID string) string {
	return fmt.Sprintf("%suser:%s:%s:permissions", s.cachePrefix, tenantID, userID)
}

// getUserRoleKey 生成用户角色缓存键
func (s *permissionCacheService) getUserRoleKey(userID, tenantID string) string {
	return fmt.Sprintf("%suser:%s:%s:roles", s.cachePrefix, tenantID, userID)
}