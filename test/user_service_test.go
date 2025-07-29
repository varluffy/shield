// Package test contains unit tests for user service.
package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varluffy/shield/internal/dto"
	"github.com/varluffy/shield/internal/models"
)

// TestUserServiceUnitTests 用户服务单元测试
func TestUserServiceUnitTests(t *testing.T) {
	// 设置测试数据库
	db, cleanup := SetupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	// 设置标准测试用户
	testUsers := SetupStandardTestUsers(db)

	// 创建测试组件
	testLogger, err := NewTestLogger()
	require.NoError(t, err)

	components := NewTestComponents(db, testLogger)

	t.Run("Test CreateUser Success", func(t *testing.T) {
		ctx := context.Background()
		
		// 设置租户上下文
		ctx = context.WithValue(ctx, "tenant_id", uint64(1))
		
		req := dto.CreateUserRequest{
			Name:     "测试用户",
			Email:    "newuser@test.com",
			Password: "password123",
			Language: "zh",
			Timezone: "Asia/Shanghai",
		}

		user, err := components.UserService.CreateUser(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, req.Name, user.Name)
		assert.Equal(t, req.Email, user.Email)
		assert.NotEmpty(t, user.UUID)
	})

	t.Run("Test CreateUser Duplicate Email", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "tenant_id", uint64(1))
		
		// 使用已存在的测试用户邮箱
		req := dto.CreateUserRequest{
			Name:     "重复邮箱用户",
			Email:    "user@tenant.test", // 这个邮箱已经存在
			Password: "password123",
		}

		user, err := components.UserService.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "已存在")
	})

	t.Run("Test CreateUser Missing TenantID", func(t *testing.T) {
		ctx := context.Background()
		// 不设置 tenant_id
		
		req := dto.CreateUserRequest{
			Name:     "无租户用户",
			Email:    "notenant@test.com",
			Password: "password123",
		}

		user, err := components.UserService.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "tenant context not found")
	})

	t.Run("Test GetUserByUUID Success", func(t *testing.T) {
		ctx := context.Background()
		
		// 获取已存在的测试用户
		testUser := testUsers["user@tenant.test"]
		require.NotNil(t, testUser, "测试用户应该存在")

		user, err := components.UserService.GetUserByUUID(ctx, testUser.UUID)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, testUser.UUID, user.UUID)
		assert.Equal(t, testUser.Email, user.Email)
	})

	t.Run("Test GetUserByUUID NotFound", func(t *testing.T) {
		ctx := context.Background()
		
		user, err := components.UserService.GetUserByUUID(ctx, "non-existent-uuid")
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("Test GetUserByEmail Success", func(t *testing.T) {
		ctx := context.Background()
		
		user, err := components.UserService.GetUserByEmail(ctx, "user@tenant.test")
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "user@tenant.test", user.Email)
	})

	t.Run("Test UpdateUserByUUID Success", func(t *testing.T) {
		ctx := context.Background()
		
		// 获取测试用户
		testUser := testUsers["user@tenant.test"]
		require.NotNil(t, testUser, "测试用户应该存在")

		req := dto.UpdateUserRequest{
			Name:     "更新后的用户名",
			Language: "en",
			Timezone: "UTC",
		}

		user, err := components.UserService.UpdateUserByUUID(ctx, testUser.UUID, req)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "更新后的用户名", user.Name)
		assert.Equal(t, "en", user.Language)
		assert.Equal(t, "UTC", user.Timezone)
	})

	t.Run("Test ListUsers Success", func(t *testing.T) {
		ctx := context.Background()
		
		filter := dto.UserFilter{
			Page:     1,
			PageSize: 10,
		}

		response, err := components.UserService.ListUsers(ctx, filter)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Greater(t, response.Total, int64(0))
		assert.NotEmpty(t, response.Users)
	})

	t.Run("Test Login with Dev Bypass Success", func(t *testing.T) {
		ctx := context.Background()
		
		req := dto.LoginRequest{
			Email:     "admin@system.test",
			Password:  "admin123",
			TenantID:  "0",
			CaptchaID: "dev-bypass",
			Answer:    "test-1234", // Using test config bypass code
		}

		response, err := components.UserService.Login(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.NotNil(t, response.User)
		assert.Equal(t, req.Email, response.User.Email)
	})

	t.Run("Test Login Wrong Password", func(t *testing.T) {
		ctx := context.Background()
		
		req := dto.LoginRequest{
			Email:     "admin@system.test",
			Password:  "wrong-password",
			TenantID:  "0",
			CaptchaID: "dev-bypass",
			Answer:    "test-1234",
		}

		response, err := components.UserService.Login(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "密码错误")
	})

	t.Run("Test Login User Not Found", func(t *testing.T) {
		ctx := context.Background()
		
		req := dto.LoginRequest{
			Email:     "nonexistent@test.com",
			Password:  "password123",
			TenantID:  "1",
			CaptchaID: "dev-bypass",
			Answer:    "test-1234",
		}

		response, err := components.UserService.Login(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "用户不存在")
	})
}

// TestUserServiceValidation 用户服务验证逻辑测试
func TestUserServiceValidation(t *testing.T) {
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

	t.Run("Test CreateUser Invalid Email", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "tenant_id", uint64(1))
		
		req := dto.CreateUserRequest{
			Name:     "测试用户",
			Email:    "invalid-email", // 无效邮箱格式
			Password: "password123",
		}

		user, err := components.UserService.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("Test CreateUser Empty Name", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "tenant_id", uint64(1))
		
		req := dto.CreateUserRequest{
			Name:     "", // 空名称
			Email:    "test@example.com",
			Password: "password123",
		}

		user, err := components.UserService.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("Test CreateUser Short Password", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "tenant_id", uint64(1))
		
		req := dto.CreateUserRequest{
			Name:     "测试用户",
			Email:    "test@example.com",
			Password: "123", // 密码太短
		}

		user, err := components.UserService.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}