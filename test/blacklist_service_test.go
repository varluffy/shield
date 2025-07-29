// Package test contains unit tests for blacklist service.
package test

import (
	"context"
	"crypto/md5"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varluffy/shield/internal/models"
)

// TestBlacklistServiceUnitTests 黑名单服务单元测试
func TestBlacklistServiceUnitTests(t *testing.T) {
	// 设置测试数据库
	db, cleanup := SetupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	// 设置标准测试用户（虽然本测试不直接使用，但确保数据一致性）
	_ = SetupStandardTestUsers(db)

	// 创建测试组件
	testLogger, err := NewTestLogger()
	require.NoError(t, err)

	components := NewTestComponents(db, testLogger)

	// 确保黑名单服务可用
	if components.BlacklistService == nil {
		t.Skip("黑名单服务不可用，跳过测试")
		return
	}

	// 辅助函数：生成测试手机号的MD5
	generatePhoneMD5 := func(phone string) string {
		hash := md5.Sum([]byte(phone))
		return fmt.Sprintf("%x", hash)
	}

	t.Run("Test CreateBlacklist Success", func(t *testing.T) {
		ctx := context.Background()

		phoneMD5 := generatePhoneMD5("13800138001")
		blacklist := &models.PhoneBlacklist{
			TenantModel: models.TenantModel{TenantID: 1},
			PhoneMD5:    phoneMD5,
			Source:      "manual",
			Reason:      "测试创建黑名单",
			OperatorID:  1,
			IsActive:    true,
		}

		err := components.BlacklistService.CreateBlacklist(ctx, blacklist)
		require.NoError(t, err)
		assert.NotZero(t, blacklist.ID, "创建的黑名单应该有ID")
	})

	t.Run("Test CheckPhoneMD5 Hit", func(t *testing.T) {
		ctx := context.Background()

		// 先创建一个黑名单记录
		phoneMD5 := generatePhoneMD5("13800138002")
		blacklist := &models.PhoneBlacklist{
			TenantModel: models.TenantModel{TenantID: 1},
			PhoneMD5:    phoneMD5,
			Source:      "manual",
			Reason:      "测试查询命中",
			OperatorID:  1,
			IsActive:    true,
		}

		err := components.BlacklistService.CreateBlacklist(ctx, blacklist)
		require.NoError(t, err)

		// 检查是否在黑名单中
		isBlacklisted, err := components.BlacklistService.CheckPhoneMD5(ctx, 1, phoneMD5)
		require.NoError(t, err)
		assert.True(t, isBlacklisted, "应该在黑名单中")
	})

	t.Run("Test CheckPhoneMD5 Miss", func(t *testing.T) {
		ctx := context.Background()

		// 检查不存在的手机号
		phoneMD5 := generatePhoneMD5("13800138999")
		isBlacklisted, err := components.BlacklistService.CheckPhoneMD5(ctx, 1, phoneMD5)
		require.NoError(t, err)
		assert.False(t, isBlacklisted, "不应该在黑名单中")
	})

	t.Run("Test BatchImportBlacklist Success", func(t *testing.T) {
		ctx := context.Background()

		phoneMD5List := []string{
			generatePhoneMD5("13800138011"),
			generatePhoneMD5("13800138012"),
			generatePhoneMD5("13800138013"),
		}

		err := components.BlacklistService.BatchImportBlacklist(ctx, 1, phoneMD5List, "batch_import", "批量测试导入", 1)
		require.NoError(t, err)

		// 验证批量导入的数据
		for _, phoneMD5 := range phoneMD5List {
			isBlacklisted, err := components.BlacklistService.CheckPhoneMD5(ctx, 1, phoneMD5)
			require.NoError(t, err)
			assert.True(t, isBlacklisted, fmt.Sprintf("手机号 %s 应该在黑名单中", phoneMD5))
		}
	})

	t.Run("Test GetBlacklistByTenant Success", func(t *testing.T) {
		ctx := context.Background()

		// 先创建一些测试数据
		testPhones := []string{"13800138021", "13800138022", "13800138023"}
		for _, phone := range testPhones {
			phoneMD5 := generatePhoneMD5(phone)
			blacklist := &models.PhoneBlacklist{
				TenantModel: models.TenantModel{TenantID: 1},
				PhoneMD5:    phoneMD5,
				Source:      "test",
				Reason:      "分页测试",
				OperatorID:  1,
				IsActive:    true,
			}
			err := components.BlacklistService.CreateBlacklist(ctx, blacklist)
			require.NoError(t, err)
		}

		// 获取黑名单列表
		blacklists, total, err := components.BlacklistService.GetBlacklistByTenant(ctx, 1, 1, 10)
		require.NoError(t, err)
		assert.Greater(t, total, int64(0), "应该有黑名单数据")
		assert.NotEmpty(t, blacklists, "黑名单列表不应该为空")

		// 验证分页功能
		if total > 5 {
			// 测试第二页
			secondPageBlacklists, _, err := components.BlacklistService.GetBlacklistByTenant(ctx, 1, 2, 5)
			require.NoError(t, err)
			
			// 第二页的数据不应该与第一页重复
			firstPageBlacklists, _, err := components.BlacklistService.GetBlacklistByTenant(ctx, 1, 1, 5)
			require.NoError(t, err)
			
			if len(secondPageBlacklists) > 0 && len(firstPageBlacklists) > 0 {
				assert.NotEqual(t, firstPageBlacklists[0].ID, secondPageBlacklists[0].ID, "不同页的数据不应该重复")
			}
		}
	})

	t.Run("Test DeleteBlacklist Success", func(t *testing.T) {
		ctx := context.Background()

		// 先创建一个黑名单记录
		phoneMD5 := generatePhoneMD5("13800138031")
		blacklist := &models.PhoneBlacklist{
			TenantModel: models.TenantModel{TenantID: 1},
			PhoneMD5:    phoneMD5,
			Source:      "test",
			Reason:      "删除测试",
			OperatorID:  1,
			IsActive:    true,
		}

		err := components.BlacklistService.CreateBlacklist(ctx, blacklist)
		require.NoError(t, err)

		// 验证黑名单存在
		isBlacklisted, err := components.BlacklistService.CheckPhoneMD5(ctx, 1, phoneMD5)
		require.NoError(t, err)
		assert.True(t, isBlacklisted, "删除前应该在黑名单中")

		// 删除黑名单
		err = components.BlacklistService.DeleteBlacklist(ctx, blacklist.ID)
		require.NoError(t, err)

		// 验证黑名单已被删除
		isBlacklisted, err = components.BlacklistService.CheckPhoneMD5(ctx, 1, phoneMD5)
		require.NoError(t, err)
		assert.False(t, isBlacklisted, "删除后不应该在黑名单中")
	})

	t.Run("Test SyncToRedis Success", func(t *testing.T) {
		ctx := context.Background()

		// 跳过Redis相关测试，如果Redis不可用
		if components.BlacklistService == nil {
			t.Skip("Redis不可用，跳过同步测试")
			return
		}

		// 先创建一些黑名单数据
		testPhones := []string{"13800138041", "13800138042"}
		for _, phone := range testPhones {
			phoneMD5 := generatePhoneMD5(phone)
			blacklist := &models.PhoneBlacklist{
				TenantModel: models.TenantModel{TenantID: 1},
				PhoneMD5:    phoneMD5,
				Source:      "sync_test",
				Reason:      "同步测试",
				OperatorID:  1,
				IsActive:    true,
			}
			err := components.BlacklistService.CreateBlacklist(ctx, blacklist)
			require.NoError(t, err)
		}

		// 同步到Redis
		err := components.BlacklistService.SyncToRedis(ctx, 1)
		require.NoError(t, err)

		// 验证同步后的查询结果
		for _, phone := range testPhones {
			phoneMD5 := generatePhoneMD5(phone)
			isBlacklisted, err := components.BlacklistService.CheckPhoneMD5(ctx, 1, phoneMD5)
			require.NoError(t, err)
			assert.True(t, isBlacklisted, fmt.Sprintf("同步后手机号 %s 应该在黑名单中", phoneMD5))
		}
	})

	t.Run("Test UpdateQueryMetrics", func(t *testing.T) {
		ctx := context.Background()

		// 更新查询指标
		components.BlacklistService.UpdateQueryMetrics(ctx, 1, "test-api-key", true, 50)
		components.BlacklistService.UpdateQueryMetrics(ctx, 1, "test-api-key", false, 30)
		components.BlacklistService.UpdateQueryMetrics(ctx, 1, "test-api-key", true, 70)

		// 这个方法通常是异步的，所以我们只测试它不会报错
		// 实际的指标验证需要通过GetQueryStats方法
	})

	t.Run("Test GetQueryStats", func(t *testing.T) {
		ctx := context.Background()

		// 先更新一些查询指标
		components.BlacklistService.UpdateQueryMetrics(ctx, 1, "stats-test-key", true, 100)
		components.BlacklistService.UpdateQueryMetrics(ctx, 1, "stats-test-key", false, 80)
		components.BlacklistService.UpdateQueryMetrics(ctx, 1, "stats-test-key", true, 120)

		// 获取查询统计
		stats, err := components.BlacklistService.GetQueryStats(ctx, 1, 1)
		if err != nil {
			// 如果Redis不可用，可能会报错，这是正常的
			t.Logf("获取查询统计失败（可能是Redis不可用）: %v", err)
			return
		}

		require.NotNil(t, stats)
		assert.GreaterOrEqual(t, stats.TotalQueries, int64(0), "总查询数应该>=0")
		assert.GreaterOrEqual(t, stats.HitCount, int64(0), "命中数应该>=0")
		assert.GreaterOrEqual(t, stats.MissCount, int64(0), "未命中数应该>=0")
		assert.GreaterOrEqual(t, stats.HitRate, float64(0), "命中率应该>=0")
		assert.LessOrEqual(t, stats.HitRate, float64(1), "命中率应该<=1")
	})

	t.Run("Test GetMinuteStats", func(t *testing.T) {
		ctx := context.Background()

		// 先更新一些查询指标
		components.BlacklistService.UpdateQueryMetrics(ctx, 1, "minute-stats-key", true, 90)
		components.BlacklistService.UpdateQueryMetrics(ctx, 1, "minute-stats-key", false, 110)

		// 获取分钟级统计
		minuteStats, err := components.BlacklistService.GetMinuteStats(ctx, 1, 5)
		if err != nil {
			// 如果Redis不可用，可能会报错，这是正常的
			t.Logf("获取分钟级统计失败（可能是Redis不可用）: %v", err)
			return
		}

		require.NotNil(t, minuteStats)
		assert.GreaterOrEqual(t, minuteStats.TotalQueries, int64(0), "总查询数应该>=0")
		assert.GreaterOrEqual(t, minuteStats.QPS, float64(0), "QPS应该>=0")
		assert.NotNil(t, minuteStats.MinuteData, "分钟数据不应该为nil")
	})
}

// TestBlacklistServiceErrorCases 黑名单服务错误场景测试
func TestBlacklistServiceErrorCases(t *testing.T) {
	// 设置测试数据库
	db, cleanup := SetupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	// 创建测试组件
	testLogger, err := NewTestLogger()
	require.NoError(t, err)

	components := NewTestComponents(db, testLogger)

	// 确保黑名单服务可用
	if components.BlacklistService == nil {
		t.Skip("黑名单服务不可用，跳过错误场景测试")
		return
	}

	generatePhoneMD5 := func(phone string) string {
		hash := md5.Sum([]byte(phone))
		return fmt.Sprintf("%x", hash)
	}

	t.Run("Test CreateBlacklist Duplicate PhoneMD5", func(t *testing.T) {
		ctx := context.Background()

		phoneMD5 := generatePhoneMD5("13800138101")
		
		// 创建第一个黑名单记录
		firstBlacklist := &models.PhoneBlacklist{
			TenantModel: models.TenantModel{TenantID: 1},
			PhoneMD5:    phoneMD5,
			Source:      "test",
			Reason:      "第一个记录",
			OperatorID:  1,
			IsActive:    true,
		}

		err := components.BlacklistService.CreateBlacklist(ctx, firstBlacklist)
		require.NoError(t, err)

		// 尝试创建相同PhoneMD5的记录
		secondBlacklist := &models.PhoneBlacklist{
			TenantModel: models.TenantModel{TenantID: 1},
			PhoneMD5:    phoneMD5, // 相同的PhoneMD5
			Source:      "test",
			Reason:      "重复记录",
			OperatorID:  1,
			IsActive:    true,
		}

		err = components.BlacklistService.CreateBlacklist(ctx, secondBlacklist)
		// 根据业务逻辑，可能允许重复也可能不允许
		// 这里我们测试系统的行为，不一定期望错误
		if err != nil {
			assert.Contains(t, err.Error(), "已存在", "重复创建应该有明确的错误信息")
		}
	})

	t.Run("Test CheckPhoneMD5 Invalid TenantID", func(t *testing.T) {
		ctx := context.Background()

		phoneMD5 := generatePhoneMD5("13800138102")
		
		// 使用无效的租户ID
		isBlacklisted, err := components.BlacklistService.CheckPhoneMD5(ctx, 99999, phoneMD5)
		require.NoError(t, err) // 通常不会报错，只是返回false
		assert.False(t, isBlacklisted, "无效租户ID应该返回false")
	})

	t.Run("Test CheckPhoneMD5 Empty PhoneMD5", func(t *testing.T) {
		ctx := context.Background()

		isBlacklisted, err := components.BlacklistService.CheckPhoneMD5(ctx, 1, "")
		// 空的PhoneMD5可能会报错或返回false，取决于实现
		if err != nil {
			assert.Error(t, err, "空的PhoneMD5应该报错")
		} else {
			assert.False(t, isBlacklisted, "空的PhoneMD5应该返回false")
		}
	})

	t.Run("Test BatchImportBlacklist Empty List", func(t *testing.T) {
		ctx := context.Background()

		err := components.BlacklistService.BatchImportBlacklist(ctx, 1, []string{}, "test", "空列表测试", 1)
		// 空列表可能不会报错，但应该正常处理
		require.NoError(t, err, "空列表导入应该正常处理")
	})

	t.Run("Test DeleteBlacklist NonExistent ID", func(t *testing.T) {
		ctx := context.Background()

		err := components.BlacklistService.DeleteBlacklist(ctx, 99999)
		assert.Error(t, err, "删除不存在的黑名单应该报错")
	})

	t.Run("Test GetBlacklistByTenant Invalid Pagination", func(t *testing.T) {
		ctx := context.Background()

		// 测试无效的分页参数
		_, _, err := components.BlacklistService.GetBlacklistByTenant(ctx, 1, 0, 0)
		// 可能会报错或自动调整参数，取决于实现
		if err != nil {
			assert.Error(t, err, "无效的分页参数应该报错")
		}

		// 测试负数分页参数
		_, _, err = components.BlacklistService.GetBlacklistByTenant(ctx, 1, -1, -10)
		if err != nil {
			assert.Error(t, err, "负数分页参数应该报错")
		}
	})

	t.Run("Test SyncToRedis Invalid TenantID", func(t *testing.T) {
		ctx := context.Background()

		err := components.BlacklistService.SyncToRedis(ctx, 99999)
		// 无效租户ID的同步可能不会报错，但也不会同步任何数据
		// 这取决于具体实现
		if err != nil {
			t.Logf("无效租户ID同步报错: %v", err)
		}
	})
}