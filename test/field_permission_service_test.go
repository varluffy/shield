// Package test provides unit tests for the field permission service.
// Tests cover CRUD operations, validation, and business logic for field permissions.
package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/varluffy/shield/internal/models"
)

// TestFieldPermissionServiceUnitTests 字段权限服务单元测试集合
func TestFieldPermissionServiceUnitTests(t *testing.T) {
	// 设置测试数据库
	db, cleanup := SetupTestDB(t)
	if db == nil {
		return // 跳过测试如果数据库不可用
	}
	defer cleanup()

	// 设置标准测试用户
	testUsers := SetupStandardTestUsers(db)

	// 创建测试组件
	testLogger, err := NewTestLogger()
	require.NoError(t, err)
	components := NewTestComponents(db, testLogger)

	// 创建测试上下文
	ctx := context.Background()

	// 获取标准测试用户
	systemAdmin := testUsers["admin@system.test"]
	tenantAdmin := testUsers["admin@tenant.test"]
	tenantUser := testUsers["user@tenant.test"]

	// 创建测试角色
	testRole := CreateTestRole(db, 1, "test_role", "测试角色")

	t.Run("TestGetTableFields", func(t *testing.T) {
		t.Run("Success - 获取表字段权限成功", func(t *testing.T) {
			// 获取用户表字段权限
			permissions, err := components.FieldPermissionService.GetTableFields(ctx, "users")

			assert.NoError(t, err)
			assert.NotEmpty(t, permissions)

			// 验证字段权限结构
			found := false
			for _, perm := range permissions {
				if perm.FieldName == "email" {
					assert.Equal(t, "users", perm.EntityTable)
					assert.Equal(t, "邮箱", perm.FieldLabel)
					assert.Equal(t, "default", perm.DefaultValue)
					assert.True(t, perm.IsActive)
					found = true
					break
				}
			}
			assert.True(t, found, "应该找到email字段权限")
		})

		t.Run("Success - 不存在的表返回空列表", func(t *testing.T) {
			permissions, err := components.FieldPermissionService.GetTableFields(ctx, "non_existent_table")

			assert.NoError(t, err)
			assert.Empty(t, permissions)
		})
	})

	t.Run("TestGetRoleFieldPermissions", func(t *testing.T) {
		t.Run("Success - 获取角色字段权限成功", func(t *testing.T) {
			// 先配置一些角色字段权限
			fieldPermissions, err := components.FieldPermissionService.GetTableFields(ctx, "users")
			require.NoError(t, err)
			require.NotEmpty(t, fieldPermissions)

			// 为角色配置字段权限
			var roleFieldPermissions []models.RoleFieldPermission
			for _, fp := range fieldPermissions {
				permType := "default"
				if fp.FieldName == "password" {
					permType = "hidden"
				} else if fp.FieldName == "id" {
					permType = "readonly"
				}
				
				roleFieldPermissions = append(roleFieldPermissions, models.RoleFieldPermission{
					TenantID:       1,
					RoleID:         testRole.ID,
					EntityTable:    "users",
					FieldName:      fp.FieldName,
					PermissionType: permType,
				})
			}

			err = components.FieldPermissionService.UpdateRoleFieldPermissions(ctx, testRole.ID, "users", roleFieldPermissions)
			require.NoError(t, err)

			// 获取角色字段权限
			rolePermissions, err := components.FieldPermissionService.GetRoleFieldPermissions(ctx, testRole.ID, "users")

			assert.NoError(t, err)
			assert.NotEmpty(t, rolePermissions)

			// 验证权限配置
			foundPassword := false
			foundID := false
			for _, perm := range rolePermissions {
				if perm.FieldName == "password" {
					assert.Equal(t, "hidden", perm.PermissionType)
					foundPassword = true
				} else if perm.FieldName == "id" {
					assert.Equal(t, "readonly", perm.PermissionType)
					foundID = true
				}
			}
			assert.True(t, foundPassword, "应该找到password字段权限")
			assert.True(t, foundID, "应该找到id字段权限")
		})

		t.Run("Success - 未配置的角色返回空权限", func(t *testing.T) {
			// 创建新角色但不配置字段权限
			newRole := CreateTestRole(db, 1, "new_test_role", "新测试角色")

			rolePermissions, err := components.FieldPermissionService.GetRoleFieldPermissions(ctx, newRole.ID, "users")

			assert.NoError(t, err)
			// 未配置的角色应该返回空列表
			assert.Empty(t, rolePermissions)
		})

		t.Run("Error - 无效角色ID", func(t *testing.T) {
			rolePermissions, err := components.FieldPermissionService.GetRoleFieldPermissions(ctx, 99999, "users")

			assert.Error(t, err)
			assert.Nil(t, rolePermissions)
		})
	})

	t.Run("TestUpdateRoleFieldPermissions", func(t *testing.T) {
		t.Run("Success - 更新角色字段权限成功", func(t *testing.T) {
			// 准备更新数据
			updates := []models.RoleFieldPermission{
				{
					TenantID:       1,
					RoleID:         testRole.ID,
					EntityTable:    "users",
					FieldName:      "email",
					PermissionType: "readonly",
				},
				{
					TenantID:       1,
					RoleID:         testRole.ID,
					EntityTable:    "users",
					FieldName:      "password",
					PermissionType: "hidden",
				},
				{
					TenantID:       1,
					RoleID:         testRole.ID,
					EntityTable:    "users",
					FieldName:      "name",
					PermissionType: "default",
				},
			}

			err := components.FieldPermissionService.UpdateRoleFieldPermissions(ctx, testRole.ID, "users", updates)

			assert.NoError(t, err)

			// 验证更新结果
			rolePermissions, err := components.FieldPermissionService.GetRoleFieldPermissions(ctx, testRole.ID, "users")
			require.NoError(t, err)

			expectedPerms := map[string]string{
				"email":    "readonly",
				"password": "hidden",
				"name":     "default",
			}

			for _, perm := range rolePermissions {
				if expectedPerm, exists := expectedPerms[perm.FieldName]; exists {
					assert.Equal(t, expectedPerm, perm.PermissionType, "字段 %s 权限应该正确更新", perm.FieldName)
				}
			}
		})

		t.Run("Error - 无效权限值", func(t *testing.T) {
			updates := []models.RoleFieldPermission{
				{
					TenantID:       1,
					RoleID:         testRole.ID,
					EntityTable:    "users",
					FieldName:      "email",
					PermissionType: "invalid_permission",
				},
			}

			err := components.FieldPermissionService.UpdateRoleFieldPermissions(ctx, testRole.ID, "users", updates)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid permission type")
		})

		t.Run("Error - 无效角色ID", func(t *testing.T) {
			updates := []models.RoleFieldPermission{
				{
					TenantID:       1,
					RoleID:         99999,
					EntityTable:    "users",
					FieldName:      "email",
					PermissionType: "readonly",
				},  
			}

			err := components.FieldPermissionService.UpdateRoleFieldPermissions(ctx, 99999, "users", updates)

			assert.Error(t, err)
		})

		t.Run("Error - 空更新数据", func(t *testing.T) {
			updates := []models.RoleFieldPermission{}

			err := components.FieldPermissionService.UpdateRoleFieldPermissions(ctx, testRole.ID, "users", updates)

			// 空更新应该成功，不会有错误
			assert.NoError(t, err)
		})
	})

	t.Run("TestGetUserFieldPermissions", func(t *testing.T) {
		t.Run("Success - 系统管理员获取所有权限", func(t *testing.T) {
			// 系统管理员应该拥有所有字段的访问权限
			userPermissions, err := components.FieldPermissionService.GetUserFieldPermissions(ctx, fmt.Sprintf("%d", systemAdmin.ID), "0", "users")

			assert.NoError(t, err)
			assert.NotEmpty(t, userPermissions)

			// 系统管理员应该对所有字段都有default权限
			for fieldName, permType := range userPermissions {
				assert.Equal(t, "default", permType, "系统管理员对字段 %s 应该有default权限", fieldName)
			}
		})

		t.Run("Success - 租户管理员获取权限", func(t *testing.T) {
			userPermissions, err := components.FieldPermissionService.GetUserFieldPermissions(ctx, fmt.Sprintf("%d", tenantAdmin.ID), "1", "users")

			assert.NoError(t, err)
			assert.NotEmpty(t, userPermissions)

			// 验证租户管理员的权限
			for fieldName, permType := range userPermissions {
				assert.NotEmpty(t, permType)
				assert.Contains(t, []string{"default", "readonly", "hidden"}, permType, "字段 %s 权限类型应该有效", fieldName)
			}
		})

		t.Run("Success - 普通用户获取权限", func(t *testing.T) {
			userPermissions, err := components.FieldPermissionService.GetUserFieldPermissions(ctx, fmt.Sprintf("%d", tenantUser.ID), "1", "users")

			assert.NoError(t, err)
			assert.NotEmpty(t, userPermissions)

			// 验证普通用户权限
			for fieldName, permType := range userPermissions {
				assert.NotEmpty(t, permType)
				assert.Contains(t, []string{"default", "readonly", "hidden"}, permType, "字段 %s 权限类型应该有效", fieldName)
			}
		})

		t.Run("Error - 不存在的用户", func(t *testing.T) {
			userPermissions, err := components.FieldPermissionService.GetUserFieldPermissions(ctx, "99999", "1", "users")

			assert.Error(t, err)
			assert.Nil(t, userPermissions)
		})

		t.Run("Error - 无效用户ID", func(t *testing.T) {
			userPermissions, err := components.FieldPermissionService.GetUserFieldPermissions(ctx, "invalid", "1", "users")

			assert.Error(t, err)
			assert.Nil(t, userPermissions)
			assert.Contains(t, err.Error(), "invalid user ID")
		})

		t.Run("Error - 无效租户ID", func(t *testing.T) {
			userPermissions, err := components.FieldPermissionService.GetUserFieldPermissions(ctx, fmt.Sprintf("%d", tenantUser.ID), "invalid", "users")

			assert.Error(t, err)
			assert.Nil(t, userPermissions)
			assert.Contains(t, err.Error(), "invalid tenant ID")
		})
	})

	// 移除ValidateFieldPermissionValue测试，因为这个方法不在公共接口中

	t.Run("TestFieldPermissionCRUD", func(t *testing.T) {
		t.Run("Success - 字段权限完整CRUD流程", func(t *testing.T) {
			// 创建新测试角色
			crudTestRole := CreateTestRole(db, 1, "crud_test_role", "CRUD测试角色")

			// 1. 获取初始权限（应该是空的，因为没有配置）
			initialPermissions, err := components.FieldPermissionService.GetRoleFieldPermissions(ctx, crudTestRole.ID, "users")
			require.NoError(t, err)
			assert.Empty(t, initialPermissions, "新角色应该没有配置字段权限")

			// 2. 更新权限
			updates := []models.RoleFieldPermission{
				{
					TenantID:       1,
					RoleID:         crudTestRole.ID,
					EntityTable:    "users",
					FieldName:      "email",
					PermissionType: "readonly",
				},
				{
					TenantID:       1,
					RoleID:         crudTestRole.ID,
					EntityTable:    "users",
					FieldName:      "password",
					PermissionType: "hidden",
				},
				{
					TenantID:       1,
					RoleID:         crudTestRole.ID,
					EntityTable:    "users",
					FieldName:      "name",
					PermissionType: "default",
				},
			}

			err = components.FieldPermissionService.UpdateRoleFieldPermissions(ctx, crudTestRole.ID, "users", updates)
			require.NoError(t, err)

			// 3. 验证更新结果
			updatedPermissions, err := components.FieldPermissionService.GetRoleFieldPermissions(ctx, crudTestRole.ID, "users")
			require.NoError(t, err)
			assert.NotEmpty(t, updatedPermissions, "更新后应该有字段权限配置")

			expectedPerms := map[string]string{
				"email":    "readonly",
				"password": "hidden",
				"name":     "default",
			}

			for _, perm := range updatedPermissions {
				if expectedPerm, exists := expectedPerms[perm.FieldName]; exists {
					assert.Equal(t, expectedPerm, perm.PermissionType)
				}
			}

			// 4. 再次更新（测试幂等性）
			updatesV2 := []models.RoleFieldPermission{
				{
					TenantID:       1,
					RoleID:         crudTestRole.ID,
					EntityTable:    "users",
					FieldName:      "email",
					PermissionType: "hidden", // 从readonly改为hidden
				},
			}

			err = components.FieldPermissionService.UpdateRoleFieldPermissions(ctx, crudTestRole.ID, "users", updatesV2)
			require.NoError(t, err)

			// 5. 验证再次更新结果
			finalPermissions, err := components.FieldPermissionService.GetRoleFieldPermissions(ctx, crudTestRole.ID, "users")
			require.NoError(t, err)

			for _, perm := range finalPermissions {
				if perm.FieldName == "email" {
					assert.Equal(t, "hidden", perm.PermissionType)
				}
			}
		})
	})

	t.Run("TestFieldPermissionBusinessLogic", func(t *testing.T) {
		t.Run("Success - 用户字段权限查询", func(t *testing.T) {
			// 测试用户字段权限查询功能
			businessTestRole := CreateTestRole(db, 1, "business_test_role", "业务逻辑测试角色")

			// 为角色配置特定权限
			updates := []models.RoleFieldPermission{
				{
					TenantID:       1,
					RoleID:         businessTestRole.ID,
					EntityTable:    "users",
					FieldName:      "email",
					PermissionType: "readonly",
				},
				{
					TenantID:       1,
					RoleID:         businessTestRole.ID,
					EntityTable:    "users",
					FieldName:      "password",
					PermissionType: "hidden",
				},
			}

			err := components.FieldPermissionService.UpdateRoleFieldPermissions(ctx, businessTestRole.ID, "users", updates)
			require.NoError(t, err)

			// 创建用户并分配角色
			businessUser := CreateTestUser(db, 1, "business@test.com")
			
			// 分配角色给用户
			db.Create(&models.UserRole{
				UserID:    businessUser.ID, 
				RoleID:    businessTestRole.ID,
				TenantID:  1,
				GrantedBy: systemAdmin.ID,
				IsActive:  true,
			})

			// 获取用户字段权限
			userPermissions, err := components.FieldPermissionService.GetUserFieldPermissions(ctx, fmt.Sprintf("%d", businessUser.ID), "1", "users")
			require.NoError(t, err)
			assert.NotEmpty(t, userPermissions, "用户应该有字段权限")

			// 验证返回的权限格式
			for fieldName, permType := range userPermissions {
				assert.NotEmpty(t, fieldName, "字段名不应该为空")
				assert.Contains(t, []string{"default", "readonly", "hidden"}, permType, "权限类型应该有效")
			}
		})
	})
}