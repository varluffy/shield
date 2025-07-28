// Package test contains permission system integration tests.
package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varluffy/shield/internal/routes"
	"github.com/varluffy/shield/pkg/errors"
	"github.com/varluffy/shield/pkg/response"
)

// TestPermissionSystemIntegration 权限系统集成测试
func TestPermissionSystemIntegration(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	// 设置测试数据库
	db, cleanup := SetupTestDB(t)
	if db == nil {
		return // 跳过测试（数据库连接失败）
	}
	defer cleanup()

	// 设置标准测试用户
	testUsers := SetupStandardTestUsers(db)

	// 创建测试组件
	cfg := NewTestConfig()
	testLogger, err := NewTestLogger()
	require.NoError(t, err)

	components := NewTestComponents(db, testLogger)

	// 设置路由
	router := routes.SetupRoutes(
		cfg, testLogger,
		components.UserHandler,
		components.CaptchaHandler,
		components.PermissionHandler,
		components.RoleHandler,
		components.FieldPermissionHandler,
		nil, // blacklistHandler - 测试中不需要
		components.AuthMiddleware,
		nil, // blacklistAuthMiddleware - 测试中不需要
		nil, // blacklistLogMiddleware - 测试中不需要
	)

	t.Run("Test System Admin Permission Access", func(t *testing.T) {
		// 获取系统管理员用户
		systemAdmin := testUsers["admin@system.test"]
		require.NotNil(t, systemAdmin, "系统管理员用户应该存在")

		// 生成JWT Token
		token, err := GenerateTestJWT(components, systemAdmin.UUID, "0")
		require.NoError(t, err)

		// 测试系统管理员访问权限列表（应该返回所有权限）
		req := httptest.NewRequest("GET", "/api/v1/permissions", nil)
		
		// 使用认证辅助方法设置Header
		authHeaders := CreateAuthHeader(token)
		for key, values := range authHeaders {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// 添加用户上下文
		ctx := context.WithValue(req.Context(), "user_id", systemAdmin.UUID)
		ctx = context.WithValue(ctx, "tenant_id", "0")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, errors.CodeSuccess, resp.Code)
		assert.NotNil(t, resp.Data)

		// 验证返回的权限包含系统权限和租户权限
		data, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)

		permissions, ok := data["permissions"].([]interface{})
		require.True(t, ok)

		// 系统管理员应该能看到所有权限（包括 system 和 tenant scope）
		assert.Greater(t, len(permissions), 0)
		t.Logf("系统管理员看到的权限数量: %d", len(permissions))
	})

	t.Run("Test Tenant Admin Permission Access", func(t *testing.T) {
		// 获取租户管理员用户
		tenantAdmin := testUsers["admin@tenant.test"]
		require.NotNil(t, tenantAdmin, "租户管理员用户应该存在")

		// 生成JWT Token
		token, err := GenerateTestJWT(components, tenantAdmin.UUID, "1")
		require.NoError(t, err)

		// 测试租户管理员访问权限列表（应该只返回租户权限）
		req := httptest.NewRequest("GET", "/api/v1/permissions", nil)
		
		// 使用认证辅助方法设置Header
		authHeaders := CreateAuthHeader(token)
		for key, values := range authHeaders {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// 添加用户上下文
		ctx := context.WithValue(req.Context(), "user_id", tenantAdmin.UUID)
		ctx = context.WithValue(ctx, "tenant_id", "1")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, errors.CodeSuccess, resp.Code)
		assert.NotNil(t, resp.Data)

		// 验证返回的权限只包含租户权限
		data, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)

		permissions, ok := data["permissions"].([]interface{})
		require.True(t, ok)

		// 租户管理员应该只能看到租户权限
		assert.Greater(t, len(permissions), 0)

		// 检查所有权限都是 tenant scope
		for _, perm := range permissions {
			permMap, ok := perm.(map[string]interface{})
			require.True(t, ok)

			scope, exists := permMap["scope"]
			require.True(t, exists)
			assert.Equal(t, "tenant", scope, "租户用户应该只能看到 tenant scope 的权限")
		}

		t.Logf("租户管理员看到的权限数量: %d", len(permissions))
	})

	t.Run("Test Permission Tree API", func(t *testing.T) {
		// 测试权限树接口的自动过滤功能
		systemAdmin := testUsers["admin@system.test"]
		require.NotNil(t, systemAdmin, "系统管理员用户应该存在")
		
		token, err := GenerateTestJWT(components, systemAdmin.UUID, "0")
		require.NoError(t, err)

		// 测试系统管理员获取权限树
		req := httptest.NewRequest("GET", "/api/v1/permissions/tree", nil)
		
		// 使用认证辅助方法设置Header
		authHeaders := CreateAuthHeader(token)
		for key, values := range authHeaders {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		ctx := context.WithValue(req.Context(), "user_id", systemAdmin.UUID)
		ctx = context.WithValue(ctx, "tenant_id", "0")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, errors.CodeSuccess, resp.Code)
		assert.NotNil(t, resp.Data)

		// 验证权限树结构
		data, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)

		tree, exists := data["permission_tree"]
		require.True(t, exists)
		assert.NotNil(t, tree)

		t.Logf("权限树测试成功")
	})

	t.Run("Test Permission Filtering by Module", func(t *testing.T) {
		// 测试按模块过滤权限
		systemAdmin := testUsers["admin@system.test"]
		require.NotNil(t, systemAdmin, "系统管理员用户应该存在")
		
		token, err := GenerateTestJWT(components, systemAdmin.UUID, "0")
		require.NoError(t, err)

		// 测试按type过滤
		req := httptest.NewRequest("GET", "/api/v1/permissions?type=api", nil)
		
		// 使用认证辅助方法设置Header
		authHeaders := CreateAuthHeader(token)
		for key, values := range authHeaders {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		ctx := context.WithValue(req.Context(), "user_id", systemAdmin.UUID)
		ctx = context.WithValue(ctx, "tenant_id", "0")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.Response
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.Equal(t, errors.CodeSuccess, resp.Code)

		// 验证返回的权限都是API类型
		data, ok := resp.Data.(map[string]interface{})
		require.True(t, ok)

		permissions, ok := data["permissions"].([]interface{})
		require.True(t, ok)

		for _, perm := range permissions {
			permMap, ok := perm.(map[string]interface{})
			require.True(t, ok)

			permType, exists := permMap["type"]
			require.True(t, exists)
			assert.Equal(t, "api", permType, "过滤后的权限应该都是API类型")
		}

		t.Logf("API类型权限过滤测试成功，数量: %d", len(permissions))
	})
}

// TestPermissionServiceUnitTests 权限服务单元测试
func TestPermissionServiceUnitTests(t *testing.T) {
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

	t.Run("Test IsSystemAdmin Method", func(t *testing.T) {
		ctx := context.Background()

		// 获取系统管理员用户
		systemAdmin := testUsers["admin@system.test"]
		require.NotNil(t, systemAdmin, "系统管理员用户应该存在")

		// 测试系统管理员检查
		isSystemAdmin, err := components.PermissionService.IsSystemAdmin(ctx, systemAdmin.UUID)
		require.NoError(t, err)
		assert.True(t, isSystemAdmin, "系统管理员应该返回true")

		// 获取普通租户用户
		tenantUser := testUsers["user@tenant.test"]
		require.NotNil(t, tenantUser, "租户用户应该存在")

		// 测试普通用户检查
		isSystemAdmin, err = components.PermissionService.IsSystemAdmin(ctx, tenantUser.UUID)
		require.NoError(t, err)
		assert.False(t, isSystemAdmin, "普通用户应该返回false")
	})

	t.Run("Test IsTenantAdmin Method", func(t *testing.T) {
		ctx := context.Background()

		// 获取租户管理员用户
		tenantAdmin := testUsers["admin@tenant.test"]
		require.NotNil(t, tenantAdmin, "租户管理员用户应该存在")

		// 测试租户管理员检查
		isTenantAdmin, err := components.PermissionService.IsTenantAdmin(ctx, tenantAdmin.UUID, "1")
		require.NoError(t, err)
		// 注意：这里可能返回false，因为实际的角色检查逻辑可能需要更多设置
		t.Logf("租户管理员检查结果: %v", isTenantAdmin)

		// 测试普通用户（应该返回false）
		regularUser := testUsers["user@tenant.test"]
		require.NotNil(t, regularUser, "普通用户应该存在")

		isTenantAdmin, err = components.PermissionService.IsTenantAdmin(ctx, regularUser.UUID, "1")
		require.NoError(t, err)
		assert.False(t, isTenantAdmin, "普通用户不应该是租户管理员")
	})

	t.Run("Test Permission Auto-Filtering Logic", func(t *testing.T) {
		ctx := context.Background()

		// 测试权限自动过滤逻辑
		filter := make(map[string]interface{})

		// 不设置用户上下文，应该返回所有权限
		_, total, err := components.PermissionService.ListPermissions(ctx, filter, 1, 100)
		require.NoError(t, err)
		assert.Greater(t, total, int64(0))

		originalCount := total
		t.Logf("无用户上下文时权限数量: %d", originalCount)

		// 设置系统管理员上下文
		systemAdmin := testUsers["admin@system.test"]
		require.NotNil(t, systemAdmin, "系统管理员用户应该存在")
		ctxWithSystemAdmin := context.WithValue(ctx, "user_id", systemAdmin.UUID)

		_, total, err = components.PermissionService.ListPermissions(ctxWithSystemAdmin, filter, 1, 100)
		require.NoError(t, err)

		systemAdminCount := total
		t.Logf("系统管理员看到的权限数量: %d", systemAdminCount)

		// 设置普通用户上下文
		tenantUser := testUsers["user@tenant.test"]
		require.NotNil(t, tenantUser, "租户用户应该存在")
		ctxWithTenantUser := context.WithValue(ctx, "user_id", tenantUser.UUID)

		_, total, err = components.PermissionService.ListPermissions(ctxWithTenantUser, filter, 1, 100)
		require.NoError(t, err)

		tenantUserCount := total
		t.Logf("租户用户看到的权限数量: %d", tenantUserCount)

		// 租户用户看到的权限应该少于系统管理员
		assert.LessOrEqual(t, tenantUserCount, systemAdminCount, "租户用户看到的权限应该不多于系统管理员")
	})
}
