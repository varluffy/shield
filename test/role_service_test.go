// Package test contains unit tests for role service.
package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varluffy/shield/internal/models"
)

// TestRoleServiceUnitTests 角色服务单元测试
func TestRoleServiceUnitTests(t *testing.T) {
	// 设置测试数据库
	db, cleanup := SetupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	// 设置标准测试用户（本测试不直接使用，但确保数据一致性）
	_ = SetupStandardTestUsers(db)

	// 创建测试组件
	testLogger, err := NewTestLogger()
	require.NoError(t, err)

	components := NewTestComponents(db, testLogger)

	// 创建测试权限数据
	setupPermissionTestData(db)

	t.Run("Test CreateRole Success", func(t *testing.T) {
		ctx := context.Background()

		newRole := &models.Role{
			TenantModel: models.TenantModel{TenantID: 1},
			Code:        "test_role",
			Name:        "测试角色",
			Description: "这是一个测试角色",
			Type:        "custom",
			IsActive:    true,
		}

		createdRole, err := components.RoleService.CreateRole(ctx, newRole)
		require.NoError(t, err)
		assert.NotNil(t, createdRole)
		assert.NotZero(t, createdRole.ID)
		assert.Equal(t, "test_role", createdRole.Code)
		assert.Equal(t, "测试角色", createdRole.Name)
		assert.True(t, createdRole.IsActive)
	})

	t.Run("Test CreateRole Duplicate Code", func(t *testing.T) {
		ctx := context.Background()

		// 先创建一个角色
		firstRole := &models.Role{
			TenantModel: models.TenantModel{TenantID: 1},
			Code:        "duplicate_role",
			Name:        "第一个角色",
			Type:        "custom",
			IsActive:    true,
		}

		_, err := components.RoleService.CreateRole(ctx, firstRole)
		require.NoError(t, err)

		// 尝试创建相同代码的角色
		secondRole := &models.Role{
			TenantModel: models.TenantModel{TenantID: 1},
			Code:        "duplicate_role", // 相同的代码
			Name:        "第二个角色",
			Type:        "custom",
			IsActive:    true,
		}

		_, err = components.RoleService.CreateRole(ctx, secondRole)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("Test GetRoleByID Success", func(t *testing.T) {
		ctx := context.Background()

		// 先创建一个角色
		newRole := &models.Role{
			TenantModel: models.TenantModel{TenantID: 1},
			Code:        "get_by_id_role",
			Name:        "通过ID获取的角色",
			Type:        "custom",
			IsActive:    true,
		}

		createdRole, err := components.RoleService.CreateRole(ctx, newRole)
		require.NoError(t, err)

		// 通过ID获取角色
		foundRole, err := components.RoleService.GetRoleByID(ctx, createdRole.ID)
		require.NoError(t, err)
		assert.NotNil(t, foundRole)
		assert.Equal(t, createdRole.ID, foundRole.ID)
		assert.Equal(t, createdRole.Code, foundRole.Code)
		assert.Equal(t, createdRole.Name, foundRole.Name)
	})

	t.Run("Test GetRoleByID NotFound", func(t *testing.T) {
		ctx := context.Background()

		role, err := components.RoleService.GetRoleByID(ctx, 99999)
		assert.Error(t, err)
		assert.Nil(t, role)
	})

	t.Run("Test GetRoleByCode Success", func(t *testing.T) {
		ctx := context.Background()

		// 先创建一个角色
		newRole := &models.Role{
			TenantModel: models.TenantModel{TenantID: 1},
			Code:        "get_by_code_role",
			Name:        "通过代码获取的角色",
			Type:        "custom",
			IsActive:    true,
		}

		createdRole, err := components.RoleService.CreateRole(ctx, newRole)
		require.NoError(t, err)

		// 通过代码获取角色
		foundRole, err := components.RoleService.GetRoleByCode(ctx, "get_by_code_role", "1")
		require.NoError(t, err)
		assert.NotNil(t, foundRole)
		assert.Equal(t, createdRole.Code, foundRole.Code)
		assert.Equal(t, createdRole.Name, foundRole.Name)
	})

	t.Run("Test GetRoleByCode NotFound", func(t *testing.T) {
		ctx := context.Background()

		role, err := components.RoleService.GetRoleByCode(ctx, "non_existent_role", "1")
		assert.Error(t, err)
		assert.Nil(t, role)
	})

	t.Run("Test ListRoles Success", func(t *testing.T) {
		ctx := context.Background()

		// 创建几个测试角色
		testRoles := []*models.Role{
			{
				TenantModel: models.TenantModel{TenantID: 1},
				Code:        "list_role_1",
				Name:        "列表角色1",
				Type:        "custom",
				IsActive:    true,
			},
			{
				TenantModel: models.TenantModel{TenantID: 1},
				Code:        "list_role_2",
				Name:        "列表角色2",
				Type:        "custom",
				IsActive:    true,
			},
		}

		for _, role := range testRoles {
			_, err := components.RoleService.CreateRole(ctx, role)
			require.NoError(t, err)
		}

		// 获取角色列表
		roles, total, err := components.RoleService.ListRoles(ctx, "1", 1, 10)
		require.NoError(t, err)
		assert.Greater(t, total, int64(0))
		assert.NotEmpty(t, roles)

		// 验证至少包含我们创建的角色
		foundRole1 := false
		foundRole2 := false
		for _, role := range roles {
			if role.Code == "list_role_1" {
				foundRole1 = true
			}
			if role.Code == "list_role_2" {
				foundRole2 = true
			}
		}
		assert.True(t, foundRole1, "应该找到list_role_1")
		assert.True(t, foundRole2, "应该找到list_role_2")
	})

	t.Run("Test UpdateRole Success", func(t *testing.T) {
		ctx := context.Background()

		// 先创建一个角色
		newRole := &models.Role{
			TenantModel: models.TenantModel{TenantID: 1},
			Code:        "update_role",
			Name:        "更新前的角色",
			Description: "更新前的描述",
			Type:        "custom",
			IsActive:    true,
		}

		createdRole, err := components.RoleService.CreateRole(ctx, newRole)
		require.NoError(t, err)

		// 更新角色
		createdRole.Name = "更新后的角色"
		createdRole.Description = "更新后的描述"
		createdRole.IsActive = false

		updatedRole, err := components.RoleService.UpdateRole(ctx, createdRole)
		require.NoError(t, err)
		assert.Equal(t, "更新后的角色", updatedRole.Name)
		assert.Equal(t, "更新后的描述", updatedRole.Description)
		assert.False(t, updatedRole.IsActive)
	})

	t.Run("Test DeleteRole Success", func(t *testing.T) {
		ctx := context.Background()

		// 先创建一个角色
		newRole := &models.Role{
			TenantModel: models.TenantModel{TenantID: 1},
			Code:        "delete_role",
			Name:        "待删除的角色",
			Type:        "custom",
			IsActive:    true,
		}

		createdRole, err := components.RoleService.CreateRole(ctx, newRole)
		require.NoError(t, err)

		// 删除角色
		err = components.RoleService.DeleteRole(ctx, createdRole.ID)
		require.NoError(t, err)

		// 验证角色已被删除
		deletedRole, err := components.RoleService.GetRoleByID(ctx, createdRole.ID)
		assert.Error(t, err)
		assert.Nil(t, deletedRole)
	})

	t.Run("Test AssignPermissions Success", func(t *testing.T) {
		ctx := context.Background()

		// 先创建一个角色
		newRole := &models.Role{
			TenantModel: models.TenantModel{TenantID: 1},
			Code:        "permission_role",
			Name:        "权限测试角色",
			Type:        "custom",
			IsActive:    true,
		}

		createdRole, err := components.RoleService.CreateRole(ctx, newRole)
		require.NoError(t, err)

		// 获取一些权限用于分配
		permissions, _, err := components.PermissionService.ListPermissions(ctx, map[string]interface{}{"scope": "tenant"}, 1, 5)
		require.NoError(t, err)
		require.Greater(t, len(permissions), 0, "应该有租户权限可用")

		// 提取权限ID
		permissionIDs := make([]uint64, 0, len(permissions))
		for _, perm := range permissions {
			permissionIDs = append(permissionIDs, perm.ID)
		}

		// 分配权限给角色
		err = components.RoleService.AssignPermissions(ctx, createdRole.ID, permissionIDs)
		require.NoError(t, err)

		// 验证权限已分配
		rolePermissions, err := components.RoleService.GetRolePermissions(ctx, createdRole.ID)
		require.NoError(t, err)
		assert.Greater(t, len(rolePermissions), 0, "角色应该有权限")
		assert.LessOrEqual(t, len(rolePermissions), len(permissionIDs), "权限数量不应该超过分配的数量")
	})

	t.Run("Test GetRolePermissions Success", func(t *testing.T) {
		ctx := context.Background()

		// 先创建一个角色
		newRole := &models.Role{
			TenantModel: models.TenantModel{TenantID: 1},
			Code:        "get_permissions_role",
			Name:        "获取权限测试角色",
			Type:        "custom",
			IsActive:    true,
		}

		createdRole, err := components.RoleService.CreateRole(ctx, newRole)
		require.NoError(t, err)

		// 初始状态应该没有权限
		initialPermissions, err := components.RoleService.GetRolePermissions(ctx, createdRole.ID)
		require.NoError(t, err)
		assert.Empty(t, initialPermissions, "新创建的角色应该没有权限")

		// 分配一些权限
		permissions, _, err := components.PermissionService.ListPermissions(ctx, map[string]interface{}{"scope": "tenant"}, 1, 2)
		require.NoError(t, err)
		require.Greater(t, len(permissions), 0, "应该有权限可用")

		permissionIDs := []uint64{permissions[0].ID}
		err = components.RoleService.AssignPermissions(ctx, createdRole.ID, permissionIDs)
		require.NoError(t, err)

		// 再次获取权限
		rolePermissions, err := components.RoleService.GetRolePermissions(ctx, createdRole.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, len(rolePermissions), "角色应该有1个权限")
		assert.Equal(t, permissions[0].ID, rolePermissions[0].ID, "权限ID应该匹配")
	})

	t.Run("Test RemovePermission Success", func(t *testing.T) {
		ctx := context.Background()

		// 先创建一个角色
		newRole := &models.Role{
			TenantModel: models.TenantModel{TenantID: 1},
			Code:        "remove_permission_role",
			Name:        "移除权限测试角色",
			Type:        "custom",
			IsActive:    true,
		}

		createdRole, err := components.RoleService.CreateRole(ctx, newRole)
		require.NoError(t, err)

		// 获取一些权限并分配
		permissions, _, err := components.PermissionService.ListPermissions(ctx, map[string]interface{}{"scope": "tenant"}, 1, 3)
		require.NoError(t, err)
		require.Greater(t, len(permissions), 1, "需要至少2个权限进行测试")

		permissionIDs := []uint64{permissions[0].ID, permissions[1].ID}
		err = components.RoleService.AssignPermissions(ctx, createdRole.ID, permissionIDs)
		require.NoError(t, err)

		// 验证权限已分配
		rolePermissions, err := components.RoleService.GetRolePermissions(ctx, createdRole.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, len(rolePermissions), "角色应该有2个权限")

		// 移除一个权限
		err = components.RoleService.RemovePermission(ctx, createdRole.ID, permissions[0].ID)
		require.NoError(t, err)

		// 验证权限已被移除
		rolePermissions, err = components.RoleService.GetRolePermissions(ctx, createdRole.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, len(rolePermissions), "角色应该剩下1个权限")
		assert.Equal(t, permissions[1].ID, rolePermissions[0].ID, "剩下的权限应该是第二个权限")
	})
}

// TestRoleServiceErrorCases 角色服务错误场景测试
func TestRoleServiceErrorCases(t *testing.T) {
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

	t.Run("Test CreateRole Empty Code", func(t *testing.T) {
		ctx := context.Background()

		invalidRole := &models.Role{
			TenantModel: models.TenantModel{TenantID: 1},
			Code:        "", // 空代码
			Name:        "无效角色",
			Type:        "custom",
			IsActive:    true,
		}

		_, err := components.RoleService.CreateRole(ctx, invalidRole)
		assert.Error(t, err, "创建空代码的角色应该报错")
	})

	t.Run("Test UpdateRole NonExistent", func(t *testing.T) {
		ctx := context.Background()

		nonExistentRole := &models.Role{
			TenantModel: models.TenantModel{TenantID: 1, ID: 99999},
			Code:        "non_existent",
			Name:        "不存在的角色",
			Type:        "custom",
			IsActive:    true,
		}

		_, err := components.RoleService.UpdateRole(ctx, nonExistentRole)
		assert.Error(t, err, "更新不存在的角色应该报错")
	})

	t.Run("Test DeleteRole NonExistent", func(t *testing.T) {
		ctx := context.Background()

		err := components.RoleService.DeleteRole(ctx, 99999)
		assert.Error(t, err, "删除不存在的角色应该报错")
	})

	t.Run("Test AssignPermissions NonExistent Role", func(t *testing.T) {
		ctx := context.Background()

		err := components.RoleService.AssignPermissions(ctx, 99999, []uint64{1, 2, 3})
		assert.Error(t, err, "为不存在的角色分配权限应该报错")
	})

	t.Run("Test AssignPermissions NonExistent Permissions", func(t *testing.T) {
		ctx := context.Background()

		// 先创建一个角色
		newRole := &models.Role{
			TenantModel: models.TenantModel{TenantID: 1},
			Code:        "error_test_role",
			Name:        "错误测试角色",
			Type:        "custom",
			IsActive:    true,
		}

		createdRole, err := components.RoleService.CreateRole(ctx, newRole)
		require.NoError(t, err)

		// 尝试分配不存在的权限
		err = components.RoleService.AssignPermissions(ctx, createdRole.ID, []uint64{99999, 99998})
		assert.Error(t, err, "分配不存在的权限应该报错")
	})

	t.Run("Test RemovePermission NonExistent Role", func(t *testing.T) {
		ctx := context.Background()

		err := components.RoleService.RemovePermission(ctx, 99999, 1)
		assert.Error(t, err, "从不存在的角色移除权限应该报错")
	})

	t.Run("Test GetRolePermissions NonExistent Role", func(t *testing.T) {
		ctx := context.Background()

		permissions, err := components.RoleService.GetRolePermissions(ctx, 99999)
		assert.Error(t, err, "获取不存在角色的权限应该报错")
		assert.Nil(t, permissions)
	})
}