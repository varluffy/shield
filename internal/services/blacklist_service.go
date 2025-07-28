// Package services provides business logic layer implementations.
// This file contains blacklist service for phone blacklist operations.
package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/internal/repositories"
	"github.com/varluffy/shield/pkg/logger"
	redisClient "github.com/varluffy/shield/pkg/redis"
	"go.uber.org/zap"
)

// BlacklistService 黑名单服务接口
type BlacklistService interface {
	CheckPhoneMD5(ctx context.Context, tenantID uint64, phoneMD5 string) (bool, error)
	CreateBlacklist(ctx context.Context, blacklist *models.PhoneBlacklist) error
	BatchImportBlacklist(ctx context.Context, tenantID uint64, phoneMD5List []string, source, reason string, operatorID uint64) error
	GetBlacklistByTenant(ctx context.Context, tenantID uint64, page, pageSize int) ([]*models.PhoneBlacklist, int64, error)
	DeleteBlacklist(ctx context.Context, id uint64) error
	SyncToRedis(ctx context.Context, tenantID uint64) error
	GetQueryStats(ctx context.Context, tenantID uint64, hours int) (*QueryStats, error)
	GetMinuteStats(ctx context.Context, tenantID uint64, minutes int) (*MinuteStats, error)
	UpdateQueryMetrics(ctx context.Context, tenantID uint64, apiKey string, isHit bool, latencyMs int64)
}

// QueryStats 查询统计信息
type QueryStats struct {
	TotalQueries int64   `json:"total_queries"`
	HitCount     int64   `json:"hit_count"`
	MissCount    int64   `json:"miss_count"`
	HitRate      float64 `json:"hit_rate"`
	AvgLatency   float64 `json:"avg_latency_ms"`
}

// MinuteStats 分钟级统计信息
type MinuteStats struct {
	Timestamp    time.Time     `json:"timestamp"`
	TotalQueries int64         `json:"total_queries"`
	HitCount     int64         `json:"hit_count"`
	MissCount    int64         `json:"miss_count"`
	HitRate      float64       `json:"hit_rate"`
	QPS          float64       `json:"qps"`
	AvgLatency   float64       `json:"avg_latency_ms"`
	MinuteData   []MinutePoint `json:"minute_data"`
}

// MinutePoint 每分钟数据点
type MinutePoint struct {
	Minute       string  `json:"minute"`
	TotalQueries int64   `json:"total_queries"`
	HitCount     int64   `json:"hit_count"`
	QPS          float64 `json:"qps"`
	AvgLatency   float64 `json:"avg_latency_ms"`
}

// blacklistService 黑名单服务实现
type blacklistService struct {
	blacklistRepo repositories.BlacklistRepository
	redis         *redisClient.Client
	logger        *logger.Logger
}

// NewBlacklistService 创建黑名单服务
func NewBlacklistService(
	blacklistRepo repositories.BlacklistRepository,
	redis *redisClient.Client,
	logger *logger.Logger,
) BlacklistService {
	return &blacklistService{
		blacklistRepo: blacklistRepo,
		redis:         redis,
		logger:        logger,
	}
}

// CheckPhoneMD5 检查手机号MD5是否在黑名单中
func (s *blacklistService) CheckPhoneMD5(ctx context.Context, tenantID uint64, phoneMD5 string) (bool, error) {
	// 构建Redis key
	redisKey := fmt.Sprintf("blacklist:tenant:%d", tenantID)

	// 从Redis SET中检查
	exists, err := s.redis.SIsMember(ctx, redisKey, phoneMD5).Result()
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Redis查询失败，回退到数据库查询",
			zap.Error(err),
			zap.Uint64("tenant_id", tenantID),
			zap.String("phone_md5", phoneMD5))

		// Redis失败时回退到数据库查询
		return s.blacklistRepo.ExistsByTenantAndMD5(ctx, tenantID, phoneMD5)
	}

	s.logger.DebugWithTrace(ctx, "黑名单查询完成",
		zap.Uint64("tenant_id", tenantID),
		zap.String("phone_md5", phoneMD5),
		zap.Bool("is_hit", exists))

	return exists, nil
}

// CreateBlacklist 创建黑名单记录
func (s *blacklistService) CreateBlacklist(ctx context.Context, blacklist *models.PhoneBlacklist) error {
	// 创建数据库记录
	err := s.blacklistRepo.Create(ctx, blacklist)
	if err != nil {
		return fmt.Errorf("创建黑名单记录失败: %w", err)
	}

	// 同步到Redis
	redisKey := fmt.Sprintf("blacklist:tenant:%d", blacklist.TenantID)
	err = s.redis.SAdd(ctx, redisKey, blacklist.PhoneMD5).Err()
	if err != nil {
		s.logger.WarnWithTrace(ctx, "同步黑名单到Redis失败",
			zap.Error(err),
			zap.Uint64("tenant_id", blacklist.TenantID),
			zap.String("phone_md5", blacklist.PhoneMD5))
	}

	s.logger.InfoWithTrace(ctx, "黑名单记录创建成功",
		zap.Uint64("tenant_id", blacklist.TenantID),
		zap.String("phone_md5", blacklist.PhoneMD5),
		zap.String("source", blacklist.Source))

	return nil
}

// BatchImportBlacklist 批量导入黑名单
func (s *blacklistService) BatchImportBlacklist(ctx context.Context, tenantID uint64, phoneMD5List []string, source, reason string, operatorID uint64) error {
	if len(phoneMD5List) == 0 {
		return fmt.Errorf("导入列表为空")
	}

	// 准备批量数据
	blacklists := make([]*models.PhoneBlacklist, 0, len(phoneMD5List))
	redisValues := make([]interface{}, 0, len(phoneMD5List))

	for _, phoneMD5 := range phoneMD5List {
		blacklists = append(blacklists, &models.PhoneBlacklist{
			TenantModel: models.TenantModel{TenantID: tenantID},
			PhoneMD5:    phoneMD5,
			Source:      source,
			Reason:      reason,
			OperatorID:  operatorID,
			IsActive:    true,
		})
		redisValues = append(redisValues, phoneMD5)
	}

	// 批量插入数据库
	err := s.blacklistRepo.BatchCreate(ctx, blacklists)
	if err != nil {
		return fmt.Errorf("批量导入黑名单失败: %w", err)
	}

	// 批量同步到Redis
	redisKey := fmt.Sprintf("blacklist:tenant:%d", tenantID)
	err = s.redis.SAdd(ctx, redisKey, redisValues...).Err()
	if err != nil {
		s.logger.WarnWithTrace(ctx, "批量同步黑名单到Redis失败",
			zap.Error(err),
			zap.Uint64("tenant_id", tenantID),
			zap.Int("count", len(phoneMD5List)))
	}

	s.logger.InfoWithTrace(ctx, "批量导入黑名单成功",
		zap.Uint64("tenant_id", tenantID),
		zap.Int("count", len(phoneMD5List)),
		zap.String("source", source))

	return nil
}

// GetBlacklistByTenant 分页获取租户黑名单
func (s *blacklistService) GetBlacklistByTenant(ctx context.Context, tenantID uint64, page, pageSize int) ([]*models.PhoneBlacklist, int64, error) {
	offset := (page - 1) * pageSize
	return s.blacklistRepo.GetByTenant(ctx, tenantID, offset, pageSize)
}

// DeleteBlacklist 删除黑名单记录
func (s *blacklistService) DeleteBlacklist(ctx context.Context, id uint64) error {
	// 先获取记录信息（用于Redis清理）
	blacklist, err := s.blacklistRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("获取黑名单记录失败: %w", err)
	}

	// 删除数据库记录
	err = s.blacklistRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("删除黑名单记录失败: %w", err)
	}

	// 从Redis中移除
	redisKey := fmt.Sprintf("blacklist:tenant:%d", blacklist.TenantID)
	err = s.redis.SRem(ctx, redisKey, blacklist.PhoneMD5).Err()
	if err != nil {
		s.logger.WarnWithTrace(ctx, "从Redis中移除黑名单失败",
			zap.Error(err),
			zap.Uint64("tenant_id", blacklist.TenantID),
			zap.String("phone_md5", blacklist.PhoneMD5))
	}

	s.logger.InfoWithTrace(ctx, "黑名单记录删除成功",
		zap.Uint64("id", id),
		zap.Uint64("tenant_id", blacklist.TenantID),
		zap.String("phone_md5", blacklist.PhoneMD5))

	return nil
}

// SyncToRedis 同步租户黑名单到Redis
func (s *blacklistService) SyncToRedis(ctx context.Context, tenantID uint64) error {
	// 从数据库获取所有有效的MD5
	md5List, err := s.blacklistRepo.GetActiveMD5ListByTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("获取黑名单数据失败: %w", err)
	}

	redisKey := fmt.Sprintf("blacklist:tenant:%d", tenantID)

	// 清空现有Redis数据
	err = s.redis.Del(ctx, redisKey).Err()
	if err != nil {
		return fmt.Errorf("清空Redis数据失败: %w", err)
	}

	// 如果有数据，批量添加到Redis
	if len(md5List) > 0 {
		values := make([]interface{}, len(md5List))
		for i, md5 := range md5List {
			values[i] = md5
		}

		err = s.redis.SAdd(ctx, redisKey, values...).Err()
		if err != nil {
			return fmt.Errorf("同步数据到Redis失败: %w", err)
		}
	}

	s.logger.InfoWithTrace(ctx, "黑名单数据同步到Redis成功",
		zap.Uint64("tenant_id", tenantID),
		zap.Int("count", len(md5List)))

	return nil
}

// GetQueryStats 获取查询统计信息
func (s *blacklistService) GetQueryStats(ctx context.Context, tenantID uint64, hours int) (*QueryStats, error) {
	stats := &QueryStats{}

	// 从Redis获取统计数据（最近N小时）
	now := time.Now()
	totalQueries := int64(0)
	hitCount := int64(0)
	totalLatency := int64(0)

	for i := 0; i < hours; i++ {
		hour := now.Add(-time.Duration(i) * time.Hour)
		statsKey := fmt.Sprintf("stats:query:tenant:%d:%s", tenantID, hour.Format("2006010215"))

		// 获取每小时统计
		hourStats, err := s.redis.HMGet(ctx, statsKey, "total", "hits", "latency").Result()
		if err != nil {
			continue
		}

		if len(hourStats) >= 3 {
			if total, err := strconv.ParseInt(fmt.Sprintf("%v", hourStats[0]), 10, 64); err == nil {
				totalQueries += total
			}
			if hits, err := strconv.ParseInt(fmt.Sprintf("%v", hourStats[1]), 10, 64); err == nil {
				hitCount += hits
			}
			if latency, err := strconv.ParseInt(fmt.Sprintf("%v", hourStats[2]), 10, 64); err == nil {
				totalLatency += latency
			}
		}
	}

	stats.TotalQueries = totalQueries
	stats.HitCount = hitCount
	stats.MissCount = totalQueries - hitCount

	if totalQueries > 0 {
		stats.HitRate = float64(hitCount) / float64(totalQueries) * 100
		stats.AvgLatency = float64(totalLatency) / float64(totalQueries)
	}

	return stats, nil
}

// GetMinuteStats 获取分钟级统计信息
func (s *blacklistService) GetMinuteStats(ctx context.Context, tenantID uint64, minutes int) (*MinuteStats, error) {
	stats := &MinuteStats{
		Timestamp:  time.Now(),
		MinuteData: make([]MinutePoint, 0, minutes),
	}

	now := time.Now()
	totalQueries := int64(0)
	hitCount := int64(0)
	totalLatency := int64(0)

	// 获取最近N分钟的数据
	for i := 0; i < minutes; i++ {
		minute := now.Add(-time.Duration(i) * time.Minute)
		minuteKey := fmt.Sprintf("stats:minute:tenant:%d:%s", tenantID, minute.Format("200601021504"))

		// 获取每分钟统计
		minuteStats, err := s.redis.HMGet(ctx, minuteKey, "total", "hits", "latency", "count").Result()
		if err != nil {
			continue
		}

		var minuteTotal, minuteHits, minuteLatency, minuteCount int64

		if len(minuteStats) >= 4 {
			if total, err := strconv.ParseInt(fmt.Sprintf("%v", minuteStats[0]), 10, 64); err == nil {
				minuteTotal = total
				totalQueries += total
			}
			if hits, err := strconv.ParseInt(fmt.Sprintf("%v", minuteStats[1]), 10, 64); err == nil {
				minuteHits = hits
				hitCount += hits
			}
			if latency, err := strconv.ParseInt(fmt.Sprintf("%v", minuteStats[2]), 10, 64); err == nil {
				minuteLatency = latency
				totalLatency += latency
			}
			if count, err := strconv.ParseInt(fmt.Sprintf("%v", minuteStats[3]), 10, 64); err == nil {
				minuteCount = count
			}
		}

		// 计算该分钟的统计
		point := MinutePoint{
			Minute:       minute.Format("15:04"),
			TotalQueries: minuteTotal,
			HitCount:     minuteHits,
			QPS:          float64(minuteTotal) / 60.0,
		}

		if minuteCount > 0 {
			point.AvgLatency = float64(minuteLatency) / float64(minuteCount)
		}

		stats.MinuteData = append(stats.MinuteData, point)
	}

	// 计算总体统计
	stats.TotalQueries = totalQueries
	stats.HitCount = hitCount
	stats.MissCount = totalQueries - hitCount

	if totalQueries > 0 {
		stats.HitRate = float64(hitCount) / float64(totalQueries) * 100
		stats.QPS = float64(totalQueries) / float64(minutes*60)
		stats.AvgLatency = float64(totalLatency) / float64(totalQueries)
	}

	return stats, nil
}

// UpdateQueryMetrics 更新查询指标
func (s *blacklistService) UpdateQueryMetrics(ctx context.Context, tenantID uint64, apiKey string, isHit bool, latencyMs int64) {
	now := time.Now()

	// 分钟级统计key
	minuteKey := fmt.Sprintf("stats:minute:tenant:%d:%s", tenantID, now.Format("200601021504"))
	hourKey := fmt.Sprintf("stats:query:tenant:%d:%s", tenantID, now.Format("2006010215"))

	// 使用Pipeline批量更新
	pipe := s.redis.Pipeline()

	// 更新分钟级统计
	pipe.HIncrBy(ctx, minuteKey, "total", 1)
	if isHit {
		pipe.HIncrBy(ctx, minuteKey, "hits", 1)
	}
	pipe.HIncrBy(ctx, minuteKey, "latency", latencyMs)
	pipe.HIncrBy(ctx, minuteKey, "count", 1)
	pipe.Expire(ctx, minuteKey, 2*time.Hour) // 保留2小时

	// 更新小时级统计
	pipe.HIncrBy(ctx, hourKey, "total", 1)
	if isHit {
		pipe.HIncrBy(ctx, hourKey, "hits", 1)
	}
	pipe.HIncrBy(ctx, hourKey, "latency", latencyMs)
	pipe.Expire(ctx, hourKey, 48*time.Hour) // 保留48小时

	// 更新API Key的分钟级统计
	apiMinuteKey := fmt.Sprintf("stats:minute:api:%s:%s", apiKey, now.Format("200601021504"))
	pipe.HIncrBy(ctx, apiMinuteKey, "total", 1)
	if isHit {
		pipe.HIncrBy(ctx, apiMinuteKey, "hits", 1)
	}
	pipe.Expire(ctx, apiMinuteKey, 1*time.Hour) // API统计保留1小时

	_, err := pipe.Exec(ctx)
	if err != nil {
		s.logger.WarnWithTrace(ctx, "更新查询指标失败",
			zap.Uint64("tenant_id", tenantID),
			zap.String("api_key", apiKey),
			zap.Error(err))
	}
}
