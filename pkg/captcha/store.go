// Package captcha provides captcha generation and verification functionality.
package captcha

import (
	"context"
	"fmt"
	"time"

	"github.com/mojocn/base64Captcha/store"
	"github.com/varluffy/shield/pkg/redis"
	"go.uber.org/zap"
)

// RedisCaptchaStore implements store.Store interface using Redis
type RedisCaptchaStore struct {
	redisClient *redis.Client
	prefix      string
	expiration  time.Duration
	logger      *zap.Logger
}

// NewRedisCaptchaStore creates a new Redis-based captcha store
func NewRedisCaptchaStore(redisClient *redis.Client, keyPrefix string, expiration time.Duration, logger *zap.Logger) store.Store {
	return &RedisCaptchaStore{
		redisClient: redisClient,
		prefix:      keyPrefix + "captcha:",
		expiration:  expiration,
		logger:      logger,
	}
}

// Set implements store.Store interface
func (s *RedisCaptchaStore) Set(id string, value string) {
	ctx := context.Background()
	key := s.getKey(id)
	
	err := s.redisClient.Set(ctx, key, value, s.expiration).Err()
	if err != nil {
		s.logger.Error("Failed to set captcha in Redis", 
			zap.String("id", id), 
			zap.Error(err))
		return
	}
	
	s.logger.Debug("Captcha set successfully", 
		zap.String("id", id), 
		zap.String("key", key))
}

// Get implements store.Store interface
func (s *RedisCaptchaStore) Get(id string, clear bool) string {
	ctx := context.Background()
	key := s.getKey(id)
	
	value, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err.Error() != "redis: nil" {
			s.logger.Error("Failed to get captcha from Redis", 
				zap.String("id", id), 
				zap.Error(err))
		}
		return ""
	}
	
	// 如果需要清除，立即删除
	if clear {
		err = s.redisClient.Del(ctx, key).Err()
		if err != nil {
			s.logger.Error("Failed to delete captcha from Redis", 
				zap.String("id", id), 
				zap.Error(err))
		}
	}
	
	s.logger.Debug("Captcha retrieved successfully", 
		zap.String("id", id), 
		zap.String("key", key),
		zap.Bool("clear", clear))
	
	return value
}

// getKey generates the Redis key for captcha
func (s *RedisCaptchaStore) getKey(id string) string {
	return fmt.Sprintf("%s%s", s.prefix, id)
} 