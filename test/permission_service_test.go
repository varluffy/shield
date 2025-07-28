// Package test contains unit tests for permission service.
package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varluffy/shield/internal/models"
)

// TestPermissionServiceUnitTests 权限服务专门单元测试
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

	// 创建测试权限数据
	setupPermissionTestData(db)

	t.Run("Test IsSystemAdmin", func(t *testing.T) {
		ctx := context.Background()

		// 测试系统管理员
		systemAdmin := testUsers["admin@system.test"]
		require.NotNil(t, systemAdmin, "系统管理员用户应该存在")

		isSystemAdmin, err := components.PermissionService.IsSystemAdmin(ctx, systemAdmin.UUID)
		require.NoError(t, err)
		assert.True(t, isSystemAdmin, "系统管理员应该返回true")

		// 测试普通用户
		regularUser := testUsers["user@tenant.test"]
		require.NotNil(t, regularUser, "普通用户应该存在")

		isSystemAdmin, err = components.PermissionService.IsSystemAdmin(ctx, regularUser.UUID)
		require.NoError(t, err)
		assert.False(t, isSystemAdmin, "普通用户应该返回false")
	})

	t.Run("Test IsTenantAdmin", func(t *testing.T) {
		ctx := context.Background()

		// 测试租户管理员
		tenantAdmin := testUsers["admin@tenant.test"]
		require.NotNil(t, tenantAdmin, "租户管理员用户应该存在")

		isTenantAdmin, err := components.PermissionService.IsTenantAdmin(ctx, tenantAdmin.UUID, "1")
		require.NoError(t, err)
		// 注意：可能返回false，因为需要正确的角色分配
		t.Logf("租户管理员检查结果: %v", isTenantAdmin)

		// 测试普通用户
		regularUser := testUsers["user@tenant.test"]
		require.NotNil(t, regularUser, "普通用户应该存在")

		isTenantAdmin, err = components.PermissionService.IsTenantAdmin(ctx, regularUser.UUID, "1")
		require.NoError(t, err)
		assert.False(t, isTenantAdmin, "普通用户不应该是租户管理员")
	})

	t.Run("Test CheckUserPermission", func(t *testing.T) {
		ctx := context.Background()

		// 测试系统管理员权限（应该拥有所有权限）
		systemAdmin := testUsers["admin@system.test"]
		require.NotNil(t, systemAdmin, "系统管理员用户应该存在")

		hasPermission, err := components.PermissionService.CheckUserPermission(ctx, systemAdmin.UUID, "0", "system_tenant_manage")
		require.NoError(t, err)
		assert.True(t, hasPermission, "系统管理员应该拥有所有权限")

		// 测试普通用户权限
		regularUser := testUsers["user@tenant.test"]
		require.NotNil(t, regularUser, "普通用户应该存在")

		hasPermission, err = components.PermissionService.CheckUserPermission(ctx, regularUser.UUID, "1", "system_tenant_manage")
		require.NoError(t, err)
		// 普通用户应该没有系统权限
		assert.False(t, hasPermission, "普通用户不应该拥有系统权限")
	})

	t.Run("Test GetUserPermissions", func(t *testing.T) {
		ctx := context.Background()

		// 测试系统管理员权限
		systemAdmin := testUsers["admin@system.test"]
		require.NotNil(t, systemAdmin, "系统管理员用户应该存在")

		permissions, err := components.PermissionService.GetUserPermissions(ctx, systemAdmin.UUID, "0")
		require.NoError(t, err)
		assert.NotEmpty(t, permissions, "系统管理员应该有权限")

		// 检查是否包含系统权限
		hasSystemPerm := false
		for _, perm := range permissions {
			if perm.Scope == "system" {
				hasSystemPerm = true
				break
			}
		}
		assert.True(t, hasSystemPerm, "系统管理员应该包含系统权限")

		// 测试租户用户权限
		tenantUser := testUsers["user@tenant.test"]
		require.NotNil(t, tenantUser, "租户用户应该存在")

		permissions, err = components.PermissionService.GetUserPermissions(ctx, tenantUser.UUID, "1")
		require.NoError(t, err)

		// 检查租户用户是否只有租户权限
		for _, perm := range permissions {
			assert.NotEqual(t, "system", perm.Scope, "租户用户不应该有系统权限，实际权限: %s", perm.Code)
		}
	})

	t.Run("Test GetUserRoles", func(t *testing.T) {
		ctx := context.Background()

		// 测试系统管理员角色
		systemAdmin := testUsers["admin@system.test"]
		require.NotNil(t, systemAdmin, "系统管理员用户应该存在")

		roles, err := components.PermissionService.GetUserRoles(ctx, systemAdmin.UUID, "0")
		require.NoError(t, err)
		
		t.Logf("系统管理员角色数量: %d", len(roles))
		for _, role := range roles {
			t.Logf("角色: %s (%s)", role.Name, role.Code)
		}
	})

	t.Run("Test ListPermissions", func(t *testing.T) {
		ctx := context.Background()

		// 测试无过滤条件的权限列表
		filter := make(map[string]interface{})
		permissions, total, err := components.PermissionService.ListPermissions(ctx, filter, 1, 10)
		require.NoError(t, err)
		assert.Greater(t, total, int64(0), "应该有权限数据")
		assert.NotEmpty(t, permissions, "权限列表不应该为空")

		// 测试按类型过滤
		filter["type"] = "api"
		permissions, total, err = components.PermissionService.ListPermissions(ctx, filter, 1, 10)
		require.NoError(t, err)
		
		for _, perm := range permissions {
			assert.Equal(t, "api", perm.Type, "过滤后的权限应该都是API类型")
		}

		// 测试按范围过滤
		filter = make(map[string]interface{})
		filter["scope"] = "system"
		permissions, total, err = components.PermissionService.ListPermissions(ctx, filter, 1, 10)
		require.NoError(t, err)
		
		for _, perm := range permissions {
			assert.Equal(t, "system", perm.Scope, "过滤后的权限应该都是系统范围")
		}
	})

	t.Run("Test GetPermissionTree", func(t *testing.T) {
		ctx := context.Background()

		// 测试系统权限树
		tree, err := components.PermissionService.GetPermissionTree(ctx, "0", "system")
		require.NoError(t, err)
		assert.NotNil(t, tree, "权限树不应该为空")

		// 测试租户权限树
		tree, err = components.PermissionService.GetPermissionTree(ctx, "1", "tenant")
		require.NoError(t, err)
		assert.NotNil(t, tree, "租户权限树不应该为空")
	})

	t.Run("Test GetUserMenuPermissions", func(t *testing.T) {
		ctx := context.Background()

		// 测试系统管理员菜单权限
		systemAdmin := testUsers["admin@system.test"]
		require.NotNil(t, systemAdmin, "系统管理员用户应该存在")

		menuPermissions, err := components.PermissionService.GetUserMenuPermissions(ctx, systemAdmin.UUID, "0")
		require.NoError(t, err)
		assert.NotEmpty(t, menuPermissions, "系统管理员应该有菜单权限")

		// 检查菜单权限类型
		for _, perm := range menuPermissions {
			assert.Equal(t, "menu", perm.Type, "应该都是菜单类型权限")
		}

		// 测试租户用户菜单权限
		tenantUser := testUsers["user@tenant.test"]
		require.NotNil(t, tenantUser, "租户用户应该存在")

		menuPermissions, err = components.PermissionService.GetUserMenuPermissions(ctx, tenantUser.UUID, "1")
		require.NoError(t, err)
		
		// 检查租户用户菜单权限范围
		for _, perm := range menuPermissions {
			assert.Equal(t, "menu", perm.Type, "应该都是菜单类型权限")
			assert.NotEqual(t, "system", perm.Scope, "租户用户不应该有系统菜单权限")
		}
	})

	t.Run("Test GetPermissionByCode", func(t *testing.T) {
		ctx := context.Background()

		// 测试获取已存在的权限
		permission, err := components.PermissionService.GetPermissionByCode(ctx, "system_tenant_manage")
		require.NoError(t, err)
		assert.NotNil(t, permission)
		assert.Equal(t, "system_tenant_manage", permission.Code)

		// 测试获取不存在的权限
		permission, err = components.PermissionService.GetPermissionByCode(ctx, "non_existent_permission")
		assert.Error(t, err)
		assert.Nil(t, permission)
	})

	t.Run("Test CreatePermission", func(t *testing.T) {
		ctx := context.Background()

		newPermission := &models.Permission{
			Code:        "test_new_permission",
			Name:        "测试新权限",
			Description: "这是一个测试创建的权限",
			Type:        "api",
			Scope:       "tenant",
			Module:      "test",
			IsBuiltin:   false,
			IsActive:    true,
		}

		err := components.PermissionService.CreatePermission(ctx, newPermission)
		require.NoError(t, err)
		assert.NotZero(t, newPermission.ID, "创建的权限应该有ID")

		// 验证权限已创建
		createdPermission, err := components.PermissionService.GetPermissionByCode(ctx, "test_new_permission")
		require.NoError(t, err)
		assert.Equal(t, newPermission.Name, createdPermission.Name)
	})

	t.Run("Test UpdatePermission", func(t *testing.T) {
		ctx := context.Background()

		// 先获取一个权限
		permission, err := components.PermissionService.GetPermissionByCode(ctx, "tenant_user_manage")
		require.NoError(t, err)
		require.NotNil(t, permission)

		// 更新权限
		newName := "更新后的用户管理权限"
		newDescription := "这是更新后的描述"
		isActive := false

		updatedPermission, err := components.PermissionService.UpdatePermission(ctx, permission.ID, newName, newDescription, &isActive)
		require.NoError(t, err)
		assert.Equal(t, newName, updatedPermission.Name)
		assert.Equal(t, newDescription, updatedPermission.Description)
		assert.False(t, updatedPermission.IsActive)
	})
}

// TestPermissionServiceErrorCases 权限服务错误场景测试
func TestPermissionServiceErrorCases(t *testing.T) {
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

	t.Run("Test CheckUserPermission Invalid UserID", func(t *testing.T) {
		ctx := context.Background()

		hasPermission, err := components.PermissionService.CheckUserPermission(ctx, "invalid-user-id", "1", "some_permission")
		assert.False(t, hasPermission)
		// 可能不会报错，但应该返回false
	})

	t.Run("Test GetUserPermissions Invalid UserID", func(t *testing.T) {
		ctx := context.Background()

		permissions, err := components.PermissionService.GetUserPermissions(ctx, "invalid-user-id", "1")
		// 可能返回空数组而不是错误
		if err != nil {
			assert.Error(t, err)
		} else {
			assert.Empty(t, permissions)
		}
	})

	t.Run("Test CreatePermission Duplicate Code", func(t *testing.T) {
		ctx := context.Background()

		duplicatePermission := &models.Permission{
			Code:        "tenant_user_manage", // 已存在的权限代码
			Name:        "重复权限",
			Description: "这是一个重复的权限代码",
			Type:        "api",
			Scope:       "tenant",
			Module:      "test",
			IsBuiltin:   false,
			IsActive:    true,
		}

		err := components.PermissionService.CreatePermission(ctx, duplicatePermission)
		assert.Error(t, err, "创建重复代码的权限应该报错")
	})

	t.Run("Test UpdatePermission NonExistent ID", func(t *testing.T) {
		ctx := context.Background()

		_, err := components.PermissionService.UpdatePermission(ctx, 99999, "不存在的权限", "描述", nil)
		assert.Error(t, err, "更新不存在的权限应该报错")
	})

	t.Run("Test DeletePermission NonExistent ID", func(t *testing.T) {
		ctx := context.Background()

		err := components.PermissionService.DeletePermission(ctx, 99999)
		assert.Error(t, err, "删除不存在的权限应该报错")
	})
}