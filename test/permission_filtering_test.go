// Package test contains unit tests for permission filtering logic.
package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varluffy/shield/internal/models"
	"gorm.io/gorm"
)

// TestPermissionFilteringUnit 权限过滤逻辑单元测试
func TestPermissionFilteringUnit(t *testing.T) {
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

	// 创建测试数据
	setupPermissionTestData(db)

	t.Run("Test Permission Service Auto Filtering", func(t *testing.T) {
		ctx := context.Background()

		// 获取标准测试用户
		systemAdmin := testUsers["admin@system.test"]
		require.NotNil(t, systemAdmin, "系统管理员用户应该存在")
		tenantUser := testUsers["user@tenant.test"]
		require.NotNil(t, tenantUser, "租户用户应该存在")

		// 测试1: 无用户上下文 - 应该返回所有权限
		filter := make(map[string]interface{})
		permissions, total, err := components.PermissionService.ListPermissions(ctx, filter, 1, 100)
		require.NoError(t, err)

		allPermissionsCount := total
		t.Logf("无用户上下文时的权限数量: %d", allPermissionsCount)

		// 验证包含系统权限和租户权限
		hasSystemPerm := false
		hasTenantPerm := false
		for _, perm := range permissions {
			if perm.Scope == "system" {
				hasSystemPerm = true
			}
			if perm.Scope == "tenant" {
				hasTenantPerm = true
			}
		}
		assert.True(t, hasSystemPerm, "应该包含系统权限")
		assert.True(t, hasTenantPerm, "应该包含租户权限")

		// 测试2: 系统管理员上下文 - 应该返回所有权限
		ctxWithSystemAdmin := context.WithValue(ctx, "user_id", systemAdmin.UUID)
		permissions, total, err = components.PermissionService.ListPermissions(ctxWithSystemAdmin, filter, 1, 100)
		require.NoError(t, err)

		systemAdminPermCount := total
		t.Logf("系统管理员看到的权限数量: %d", systemAdminPermCount)

		// 验证系统管理员可以看到所有权限
		assert.Equal(t, allPermissionsCount, systemAdminPermCount, "系统管理员应该看到所有权限")

		// 验证包含系统权限和租户权限
		hasSystemPerm = false
		hasTenantPerm = false
		for _, perm := range permissions {
			if perm.Scope == "system" {
				hasSystemPerm = true
			}
			if perm.Scope == "tenant" {
				hasTenantPerm = true
			}
		}
		assert.True(t, hasSystemPerm, "系统管理员应该能看到系统权限")
		assert.True(t, hasTenantPerm, "系统管理员应该能看到租户权限")

		// 测试3: 租户用户上下文 - 应该只返回租户权限
		ctxWithTenantUser := context.WithValue(ctx, "user_id", tenantUser.UUID)
		permissions, total, err = components.PermissionService.ListPermissions(ctxWithTenantUser, filter, 1, 100)
		require.NoError(t, err)

		tenantUserPermCount := total
		t.Logf("租户用户看到的权限数量: %d", tenantUserPermCount)

		// 验证租户用户只能看到租户权限
		assert.Less(t, tenantUserPermCount, allPermissionsCount, "租户用户看到的权限应该少于总权限")

		// 验证所有权限都是租户权限
		for _, perm := range permissions {
			assert.Equal(t, "tenant", perm.Scope, "租户用户应该只能看到tenant权限，实际看到: %s", perm.Scope)
		}
	})

	t.Run("Test Permission Tree Auto Filtering", func(t *testing.T) {
		ctx := context.Background()

		// 获取标准测试用户
		systemAdmin := testUsers["admin@system.test"]
		require.NotNil(t, systemAdmin, "系统管理员用户应该存在")
		tenantUser := testUsers["user@tenant.test"]
		require.NotNil(t, tenantUser, "租户用户应该存在")

		// 测试系统管理员的权限树
		ctxWithSystemAdmin := context.WithValue(ctx, "user_id", systemAdmin.UUID)
		tree, err := components.PermissionService.GetPermissionTree(ctxWithSystemAdmin, "0", "")
		require.NoError(t, err)
		assert.NotNil(t, tree, "系统管理员应该能获取权限树")

		// 测试租户用户的权限树
		ctxWithTenantUser := context.WithValue(ctx, "user_id", tenantUser.UUID)
		tree, err = components.PermissionService.GetPermissionTree(ctxWithTenantUser, "1", "")
		require.NoError(t, err)
		assert.NotNil(t, tree, "租户用户应该能获取权限树")
	})

	t.Run("Test IsSystemAdmin Logic", func(t *testing.T) {
		ctx := context.Background()

		// 获取标准测试用户
		systemAdmin := testUsers["admin@system.test"]
		require.NotNil(t, systemAdmin, "系统管理员用户应该存在")
		tenantUser := testUsers["user@tenant.test"]
		require.NotNil(t, tenantUser, "租户用户应该存在")

		// 测试系统管理员检查
		isSystemAdmin, err := components.PermissionService.IsSystemAdmin(ctx, systemAdmin.UUID)
		require.NoError(t, err)
		assert.True(t, isSystemAdmin, "系统管理员检查应该返回true")

		// 测试租户用户检查
		isSystemAdmin, err = components.PermissionService.IsSystemAdmin(ctx, tenantUser.UUID)
		require.NoError(t, err)
		assert.False(t, isSystemAdmin, "租户用户检查应该返回false")
	})
}

// setupPermissionTestData 创建测试权限数据
func setupPermissionTestData(db *gorm.DB) {
	// 创建系统权限
	systemPermissions := []models.Permission{
		{
			Code:        "system_tenant_manage",
			Name:        "租户管理",
			Description: "系统级租户管理权限",
			Type:        "api",
			Scope:       "system",
			Module:      "system",
			IsBuiltin:   true,
			IsActive:    true,
		},
		{
			Code:        "system_permission_manage",
			Name:        "权限管理",
			Description: "系统级权限管理权限",
			Type:        "api",
			Scope:       "system",
			Module:      "system",
			IsBuiltin:   true,
			IsActive:    true,
		},
	}

	// 创建租户权限
	tenantPermissions := []models.Permission{
		{
			Code:        "tenant_user_manage",
			Name:        "用户管理",
			Description: "租户级用户管理权限",
			Type:        "api",
			Scope:       "tenant",
			Module:      "user",
			IsBuiltin:   true,
			IsActive:    true,
		},
		{
			Code:        "tenant_role_manage",
			Name:        "角色管理",
			Description: "租户级角色管理权限",
			Type:        "api",
			Scope:       "tenant",
			Module:      "role",
			IsBuiltin:   true,
			IsActive:    true,
		},
	}

	// 插入权限数据
	for _, perm := range systemPermissions {
		db.Create(&perm)
	}

	for _, perm := range tenantPermissions {
		db.Create(&perm)
	}
}
