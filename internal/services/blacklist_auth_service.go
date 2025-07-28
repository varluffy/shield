// Package services provides business logic layer implementations.
// This file contains blacklist authentication service for HMAC signature validation.
package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/internal/repositories"
	"github.com/varluffy/shield/pkg/logger"
	redisClient "github.com/varluffy/shield/pkg/redis"
	"go.uber.org/zap"
)

// BlacklistAuthService 黑名单鉴权服务接口
type BlacklistAuthService interface {
	ValidateHMACSignature(ctx context.Context, apiKey, timestamp, nonce, signature, body string) (*models.BlacklistApiCredential, error)
	CheckRateLimit(ctx context.Context, apiKey string) error
	RecordQueryLog(ctx context.Context, apiKey string, phoneMD5 string, isHit bool, responseTime int, clientIP, userAgent, requestID string)
	UpdateAPIKeyUsage(ctx context.Context, apiKey string) error
}

// blacklistAuthService 黑名单鉴权服务实现
type blacklistAuthService struct {
	apiCredRepo repositories.ApiCredentialRepository
	redis       *redisClient.Client
	logger      *logger.Logger
}

// NewBlacklistAuthService 创建黑名单鉴权服务
func NewBlacklistAuthService(
	apiCredRepo repositories.ApiCredentialRepository,
	redis *redisClient.Client,
	logger *logger.Logger,
) BlacklistAuthService {
	return &blacklistAuthService{
		apiCredRepo: apiCredRepo,
		redis:       redis,
		logger:      logger,
	}
}

// ValidateHMACSignature 验证HMAC签名
func (s *blacklistAuthService) ValidateHMACSignature(ctx context.Context, apiKey, timestamp, nonce, signature, body string) (*models.BlacklistApiCredential, error) {
	// 1. 获取API密钥信息（优先从缓存获取）
	credential, err := s.getAPICredentialWithCache(ctx, apiKey)
	if err != nil {
		s.logger.WarnWithTrace(ctx, "API密钥不存在或已失效",
			zap.String("api_key", apiKey),
			zap.Error(err))
		return nil, fmt.Errorf("API密钥无效")
	}

	// 2. 时间戳验证（防重放攻击）
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("时间戳格式错误")
	}

	now := time.Now().Unix()
	if abs(now-ts) > 300 { // 5分钟时间窗口
		return nil, fmt.Errorf("请求已过期")
	}

	// 3. Nonce防重放验证
	nonceKey := fmt.Sprintf("nonce:%s:%s", apiKey, nonce)
	exists, err := s.redis.Exists(ctx, nonceKey).Result()
	if err != nil {
		s.logger.WarnWithTrace(ctx, "Nonce检查失败",
			zap.String("api_key", apiKey),
			zap.String("nonce", nonce),
			zap.Error(err))
	} else if exists > 0 {
		return nil, fmt.Errorf("请求重复")
	}

	// 4. HMAC签名验证
	expectedSignature := s.generateHMACSignature(apiKey, timestamp, nonce, body, credential.APISecret)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		s.logger.WarnWithTrace(ctx, "HMAC签名验证失败",
			zap.String("api_key", apiKey),
			zap.String("expected", expectedSignature),
			zap.String("received", signature))
		return nil, fmt.Errorf("签名验证失败")
	}

	// 5. 记录Nonce（设置5分钟过期）
	err = s.redis.SetEx(ctx, nonceKey, "1", 5*time.Minute).Err()
	if err != nil {
		s.logger.WarnWithTrace(ctx, "记录Nonce失败",
			zap.String("api_key", apiKey),
			zap.String("nonce", nonce),
			zap.Error(err))
	}

	s.logger.DebugWithTrace(ctx, "HMAC签名验证成功",
		zap.String("api_key", apiKey),
		zap.Uint64("tenant_id", credential.TenantID))

	return credential, nil
}

// CheckRateLimit 检查速率限制
func (s *blacklistAuthService) CheckRateLimit(ctx context.Context, apiKey string) error {
	// 获取API密钥配置（优先从缓存获取）
	credential, err := s.getAPICredentialWithCache(ctx, apiKey)
	if err != nil {
		return fmt.Errorf("获取API密钥配置失败: %w", err)
	}

	// 使用滑动窗口计数器实现速率限制
	rateLimitKey := fmt.Sprintf("rate_limit:%s", apiKey)
	current := time.Now().Unix()

	// 使用Redis Pipeline提高性能
	pipe := s.redis.Pipeline()

	// 移除1秒前的记录
	pipe.ZRemRangeByScore(ctx, rateLimitKey, "0", strconv.FormatInt(current-1, 10))

	// 添加当前请求
	pipe.ZAdd(ctx, rateLimitKey, redis.Z{
		Score:  float64(current),
		Member: strconv.FormatInt(current, 10),
	})

	// 获取当前1秒内的请求数
	pipe.ZCard(ctx, rateLimitKey)

	// 设置过期时间
	pipe.Expire(ctx, rateLimitKey, 2*time.Second)

	results, err := pipe.Exec(ctx)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "速率限制检查失败",
			zap.String("api_key", apiKey),
			zap.Error(err))
		return nil // 失败时允许通过，避免影响业务
	}

	// 获取当前请求数
	if len(results) >= 3 {
		if count, ok := results[2].(*redis.IntCmd); ok {
			currentRequests := count.Val()
			if currentRequests > int64(credential.RateLimit) {
				return fmt.Errorf("请求频率超限，限制: %d/秒，当前: %d/秒",
					credential.RateLimit, currentRequests)
			}
		}
	}

	return nil
}

// RecordQueryLog 记录查询日志（异步采样）
func (s *blacklistAuthService) RecordQueryLog(ctx context.Context, apiKey string, phoneMD5 string, isHit bool, responseTime int, clientIP, userAgent, requestID string) {
	// 异步记录，不阻塞主流程
	go func() {
		// 更新实时统计
		s.updateRealTimeStats(context.Background(), apiKey, isHit, responseTime)

		// 这里可以根据采样率决定是否记录详细日志
		// 暂时只记录统计信息，详细日志可以后续添加
	}()
}

// UpdateAPIKeyUsage 更新API密钥使用时间
func (s *blacklistAuthService) UpdateAPIKeyUsage(ctx context.Context, apiKey string) error {
	// 异步更新，不阻塞主流程
	go func() {
		err := s.apiCredRepo.UpdateLastUsedAt(context.Background(), apiKey)
		if err != nil {
			s.logger.WarnWithTrace(context.Background(), "更新API密钥使用时间失败",
				zap.String("api_key", apiKey),
				zap.Error(err))
		}
	}()
	return nil
}

// generateHMACSignature 生成HMAC签名
func (s *blacklistAuthService) generateHMACSignature(apiKey, timestamp, nonce, body, secret string) string {
	// 签名字符串格式: apiKey + timestamp + nonce + body
	message := apiKey + timestamp + nonce + body

	// 使用HMAC-SHA256生成签名
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// updateRealTimeStats 更新实时统计信息
func (s *blacklistAuthService) updateRealTimeStats(ctx context.Context, apiKey string, isHit bool, responseTime int) {
	// 按小时聚合统计
	now := time.Now()
	statsKey := fmt.Sprintf("stats:query:%s:%s", apiKey, now.Format("2006010215"))

	// 使用Redis Pipeline批量更新
	pipe := s.redis.Pipeline()

	// 总请求数+1
	pipe.HIncrBy(ctx, statsKey, "total", 1)

	// 命中数
	if isHit {
		pipe.HIncrBy(ctx, statsKey, "hits", 1)
	}

	// 累计响应时间
	pipe.HIncrBy(ctx, statsKey, "latency", int64(responseTime))

	// 设置过期时间（保留48小时）
	pipe.Expire(ctx, statsKey, 48*time.Hour)

	_, err := pipe.Exec(ctx)
	if err != nil {
		s.logger.WarnWithTrace(ctx, "更新实时统计失败",
			zap.String("api_key", apiKey),
			zap.Error(err))
	}
}

// getAPICredentialWithCache 获取API密钥信息（带缓存）
func (s *blacklistAuthService) getAPICredentialWithCache(ctx context.Context, apiKey string) (*models.BlacklistApiCredential, error) {
	// 缓存key
	cacheKey := fmt.Sprintf("api_credential:%s", apiKey)

	// 1. 尝试从缓存获取
	credentialJSON, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil && credentialJSON != "" {
		// 缓存命中，反序列化
		var credential models.BlacklistApiCredential
		if err := json.Unmarshal([]byte(credentialJSON), &credential); err == nil {
			s.logger.DebugWithTrace(ctx, "API密钥缓存命中",
				zap.String("api_key", apiKey))
			return &credential, nil
		}
	}

	// 2. 缓存未命中，从数据库查询
	credential, err := s.apiCredRepo.GetActiveByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	// 3. 将结果写入缓存（设置5分钟过期）
	credentialBytes, err := json.Marshal(credential)
	if err == nil {
		// 缓存5分钟，避免频繁查询数据库
		err = s.redis.SetEx(ctx, cacheKey, string(credentialBytes), 5*time.Minute).Err()
		if err != nil {
			s.logger.WarnWithTrace(ctx, "缓存API密钥信息失败",
				zap.String("api_key", apiKey),
				zap.Error(err))
			// 缓存失败不影响业务
		}
	}

	return credential, nil
}

// invalidateAPICredentialCache 清除API密钥缓存
func (s *blacklistAuthService) invalidateAPICredentialCache(ctx context.Context, apiKey string) {
	cacheKey := fmt.Sprintf("api_credential:%s", apiKey)
	err := s.redis.Del(ctx, cacheKey).Err()
	if err != nil {
		s.logger.WarnWithTrace(ctx, "清除API密钥缓存失败",
			zap.String("api_key", apiKey),
			zap.Error(err))
	}
}

// abs 计算绝对值
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
