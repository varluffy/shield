package main

import (
	"fmt"
	"log"

	"github.com/varluffy/shield/internal/models"
	"gorm.io/gorm"
)

// PermissionData 权限数据结构
type PermissionData struct {
	Code         string
	Name         string
	Description  string
	Type         string
	Scope        string
	ParentCode   string
	ResourcePath string
	Method       string
	SortOrder    int
	Module       string
}

// InitSystemPermissions 初始化系统权限
func InitSystemPermissions(db *gorm.DB) error {
	permissions := getSystemPermissions()

	for _, permData := range permissions {
		var existingPerm models.Permission
		result := db.Where("code = ?", permData.Code).First(&existingPerm)

		if result.Error == gorm.ErrRecordNotFound {
			// 权限不存在，创建新权限
			perm := models.Permission{
				Code:         permData.Code,
				Name:         permData.Name,
				Description:  permData.Description,
				Type:         permData.Type,
				Scope:        permData.Scope,
				ParentCode:   permData.ParentCode,
				ResourcePath: permData.ResourcePath,
				Method:       permData.Method,
				SortOrder:    permData.SortOrder,
				IsBuiltin:    true,
				IsActive:     true,
				Module:       permData.Module,
			}

			if err := db.Create(&perm).Error; err != nil {
				return fmt.Errorf("创建权限失败 %s: %v", permData.Code, err)
			}
			log.Printf("创建权限: %s - %s", permData.Code, permData.Name)
		} else if result.Error != nil {
			return fmt.Errorf("查询权限失败 %s: %v", permData.Code, result.Error)
		} else {
			// 权限已存在，更新基本信息（保持is_active状态不变）
			updates := map[string]interface{}{
				"name":          permData.Name,
				"description":   permData.Description,
				"type":          permData.Type,
				"scope":         permData.Scope,
				"parent_code":   permData.ParentCode,
				"resource_path": permData.ResourcePath,
				"method":        permData.Method,
				"sort_order":    permData.SortOrder,
				"module":        permData.Module,
			}

			if err := db.Model(&existingPerm).Updates(updates).Error; err != nil {
				return fmt.Errorf("更新权限失败 %s: %v", permData.Code, err)
			}
			log.Printf("更新权限: %s - %s", permData.Code, permData.Name)
		}
	}

	return nil
}

// getSystemPermissions 获取系统权限定义
func getSystemPermissions() []PermissionData {
	return []PermissionData{
		// 系统管理模块
		{
			Code:        "system_menu",
			Name:        "系统管理",
			Description: "系统管理菜单访问权限",
			Type:        models.PermissionTypeMenu,
			Scope:       models.ScopeSystem,
			ParentCode:  "",
			SortOrder:   1000,
			Module:      models.ModuleSystem,
		},

		// 租户管理
		{
			Code:        "tenant_menu",
			Name:        "租户管理",
			Description: "租户管理菜单访问权限",
			Type:        models.PermissionTypeMenu,
			Scope:       models.ScopeSystem,
			ParentCode:  "system_menu",
			SortOrder:   1010,
			Module:      models.ModuleTenant,
		},
		{
			Code:        "tenant_list_btn",
			Name:        "租户列表",
			Description: "查看租户列表按钮权限",
			Type:        models.PermissionTypeButton,
			Scope:       models.ScopeSystem,
			ParentCode:  "tenant_menu",
			SortOrder:   1011,
			Module:      models.ModuleTenant,
		},
		{
			Code:         "tenant_list_api",
			Name:         "租户列表API",
			Description:  "获取租户列表API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeSystem,
			ParentCode:   "tenant_list_btn",
			ResourcePath: "/api/v1/system/tenants",
			Method:       "GET",
			SortOrder:    1012,
			Module:       models.ModuleTenant,
		},
		{
			Code:        "tenant_create_btn",
			Name:        "创建租户",
			Description: "创建租户按钮权限",
			Type:        models.PermissionTypeButton,
			Scope:       models.ScopeSystem,
			ParentCode:  "tenant_menu",
			SortOrder:   1013,
			Module:      models.ModuleTenant,
		},
		{
			Code:         "tenant_create_api",
			Name:         "创建租户API",
			Description:  "创建租户API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeSystem,
			ParentCode:   "tenant_create_btn",
			ResourcePath: "/api/v1/system/tenants",
			Method:       "POST",
			SortOrder:    1014,
			Module:       models.ModuleTenant,
		},
		{
			Code:        "tenant_update_btn",
			Name:        "编辑租户",
			Description: "编辑租户按钮权限",
			Type:        models.PermissionTypeButton,
			Scope:       models.ScopeSystem,
			ParentCode:  "tenant_menu",
			SortOrder:   1015,
			Module:      models.ModuleTenant,
		},
		{
			Code:         "tenant_update_api",
			Name:         "更新租户API",
			Description:  "更新租户API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeSystem,
			ParentCode:   "tenant_update_btn",
			ResourcePath: "/api/v1/system/tenants/:id",
			Method:       "PUT",
			SortOrder:    1016,
			Module:       models.ModuleTenant,
		},
		{
			Code:        "tenant_delete_btn",
			Name:        "删除租户",
			Description: "删除租户按钮权限",
			Type:        models.PermissionTypeButton,
			Scope:       models.ScopeSystem,
			ParentCode:  "tenant_menu",
			SortOrder:   1017,
			Module:      models.ModuleTenant,
		},
		{
			Code:         "tenant_delete_api",
			Name:         "删除租户API",
			Description:  "删除租户API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeSystem,
			ParentCode:   "tenant_delete_btn",
			ResourcePath: "/api/v1/system/tenants/:id",
			Method:       "DELETE",
			SortOrder:    1018,
			Module:       models.ModuleTenant,
		},

		// 系统权限管理
		{
			Code:        "permission_menu",
			Name:        "权限管理",
			Description: "权限管理菜单访问权限",
			Type:        models.PermissionTypeMenu,
			Scope:       models.ScopeSystem,
			ParentCode:  "system_menu",
			SortOrder:   1020,
			Module:      models.ModuleSystem,
		},
		{
			Code:         "permission_list_api",
			Name:         "权限列表API",
			Description:  "获取权限列表API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeSystem,
			ParentCode:   "permission_menu",
			ResourcePath: "/api/v1/system/permissions",
			Method:       "GET",
			SortOrder:    1021,
			Module:       models.ModuleSystem,
		},
		{
			Code:         "permission_update_api",
			Name:         "更新权限API",
			Description:  "更新权限API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeSystem,
			ParentCode:   "permission_menu",
			ResourcePath: "/api/v1/system/permissions/:id",
			Method:       "PUT",
			SortOrder:    1022,
			Module:       models.ModuleSystem,
		},

		// 租户权限
		// 用户管理模块
		{
			Code:        "user_menu",
			Name:        "用户管理",
			Description: "用户管理菜单访问权限",
			Type:        models.PermissionTypeMenu,
			Scope:       models.ScopeTenant,
			ParentCode:  "",
			SortOrder:   2000,
			Module:      models.ModuleUser,
		},
		{
			Code:        "user_list_btn",
			Name:        "用户列表",
			Description: "查看用户列表按钮权限",
			Type:        models.PermissionTypeButton,
			Scope:       models.ScopeTenant,
			ParentCode:  "user_menu",
			SortOrder:   2010,
			Module:      models.ModuleUser,
		},
		{
			Code:         "user_list_api",
			Name:         "用户列表API",
			Description:  "获取用户列表API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "user_list_btn",
			ResourcePath: "/api/v1/users",
			Method:       "GET",
			SortOrder:    2011,
			Module:       models.ModuleUser,
		},
		{
			Code:        "user_create_btn",
			Name:        "创建用户",
			Description: "创建用户按钮权限",
			Type:        models.PermissionTypeButton,
			Scope:       models.ScopeTenant,
			ParentCode:  "user_menu",
			SortOrder:   2020,
			Module:      models.ModuleUser,
		},
		{
			Code:         "user_create_api",
			Name:         "创建用户API",
			Description:  "创建用户API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "user_create_btn",
			ResourcePath: "/api/v1/users",
			Method:       "POST",
			SortOrder:    2021,
			Module:       models.ModuleUser,
		},
		{
			Code:        "user_update_btn",
			Name:        "编辑用户",
			Description: "编辑用户按钮权限",
			Type:        models.PermissionTypeButton,
			Scope:       models.ScopeTenant,
			ParentCode:  "user_menu",
			SortOrder:   2030,
			Module:      models.ModuleUser,
		},
		{
			Code:         "user_update_api",
			Name:         "更新用户API",
			Description:  "更新用户API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "user_update_btn",
			ResourcePath: "/api/v1/users/:id",
			Method:       "PUT",
			SortOrder:    2031,
			Module:       models.ModuleUser,
		},
		{
			Code:        "user_delete_btn",
			Name:        "删除用户",
			Description: "删除用户按钮权限",
			Type:        models.PermissionTypeButton,
			Scope:       models.ScopeTenant,
			ParentCode:  "user_menu",
			SortOrder:   2040,
			Module:      models.ModuleUser,
		},
		{
			Code:         "user_delete_api",
			Name:         "删除用户API",
			Description:  "删除用户API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "user_delete_btn",
			ResourcePath: "/api/v1/users/:id",
			Method:       "DELETE",
			SortOrder:    2041,
			Module:       models.ModuleUser,
		},

		// 角色管理模块
		{
			Code:        "role_menu",
			Name:        "角色管理",
			Description: "角色管理菜单访问权限",
			Type:        models.PermissionTypeMenu,
			Scope:       models.ScopeTenant,
			ParentCode:  "",
			SortOrder:   3000,
			Module:      models.ModuleRole,
		},
		{
			Code:        "role_list_btn",
			Name:        "角色列表",
			Description: "查看角色列表按钮权限",
			Type:        models.PermissionTypeButton,
			Scope:       models.ScopeTenant,
			ParentCode:  "role_menu",
			SortOrder:   3010,
			Module:      models.ModuleRole,
		},
		{
			Code:         "role_list_api",
			Name:         "角色列表API",
			Description:  "获取角色列表API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "role_list_btn",
			ResourcePath: "/api/v1/roles",
			Method:       "GET",
			SortOrder:    3011,
			Module:       models.ModuleRole,
		},
		{
			Code:        "role_create_btn",
			Name:        "创建角色",
			Description: "创建角色按钮权限",
			Type:        models.PermissionTypeButton,
			Scope:       models.ScopeTenant,
			ParentCode:  "role_menu",
			SortOrder:   3020,
			Module:      models.ModuleRole,
		},
		{
			Code:         "role_create_api",
			Name:         "创建角色API",
			Description:  "创建角色API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "role_create_btn",
			ResourcePath: "/api/v1/roles",
			Method:       "POST",
			SortOrder:    3021,
			Module:       models.ModuleRole,
		},
		{
			Code:        "role_assign_btn",
			Name:        "分配权限",
			Description: "分配权限给角色按钮权限",
			Type:        models.PermissionTypeButton,
			Scope:       models.ScopeTenant,
			ParentCode:  "role_menu",
			SortOrder:   3030,
			Module:      models.ModuleRole,
		},
		{
			Code:         "role_assign_api",
			Name:         "分配权限API",
			Description:  "分配权限给角色API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "role_assign_btn",
			ResourcePath: "/api/v1/roles/:id/permissions",
			Method:       "POST",
			SortOrder:    3031,
			Module:       models.ModuleRole,
		},
		{
			Code:        "field_permission_btn",
			Name:        "字段权限配置",
			Description: "配置字段权限按钮权限",
			Type:        models.PermissionTypeButton,
			Scope:       models.ScopeTenant,
			ParentCode:  "role_menu",
			SortOrder:   3040,
			Module:      models.ModuleRole,
		},
		{
			Code:         "field_permission_list_api",
			Name:         "字段权限列表API",
			Description:  "获取字段权限配置API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "field_permission_btn",
			ResourcePath: "/api/v1/roles/:id/field-permissions/:table",
			Method:       "GET",
			SortOrder:    3041,
			Module:       models.ModuleRole,
		},
		{
			Code:         "field_permission_update_api",
			Name:         "更新字段权限API",
			Description:  "更新字段权限配置API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "field_permission_btn",
			ResourcePath: "/api/v1/roles/:id/field-permissions/:table",
			Method:       "PUT",
			SortOrder:    3042,
			Module:       models.ModuleRole,
		},
	}
}

// InitSystemRoles 初始化系统角色
func InitSystemRoles(db *gorm.DB) error {
	roles := []struct {
		Code        string
		Name        string
		Description string
		Type        string
		Permissions []string
	}{
		{
			Code:        models.RoleSystemAdmin,
			Name:        "系统管理员",
			Description: "系统管理员，拥有所有系统权限",
			Type:        models.RoleTypeSystem,
			Permissions: []string{
				"system_menu", "tenant_menu", "tenant_list_btn", "tenant_list_api",
				"tenant_create_btn", "tenant_create_api", "tenant_update_btn", "tenant_update_api",
				"tenant_delete_btn", "tenant_delete_api", "permission_menu", "permission_list_api",
				"permission_update_api",
			},
		},
		{
			Code:        models.RoleTenantAdmin,
			Name:        "租户管理员",
			Description: "租户管理员，拥有租户内所有权限",
			Type:        models.RoleTypeSystem,
			Permissions: []string{
				"user_menu", "user_list_btn", "user_list_api", "user_create_btn", "user_create_api",
				"user_update_btn", "user_update_api", "user_delete_btn", "user_delete_api",
				"role_menu", "role_list_btn", "role_list_api", "role_create_btn", "role_create_api",
				"role_assign_btn", "role_assign_api", "field_permission_btn", "field_permission_list_api",
				"field_permission_update_api",
			},
		},
	}

	for _, roleData := range roles {
		// 注意：系统角色不绑定租户，使用TenantID=0
		var existingRole models.Role
		result := db.Where("code = ? AND tenant_id = 0", roleData.Code).First(&existingRole)

		if result.Error == gorm.ErrRecordNotFound {
			// 角色不存在，创建新角色
			role := models.Role{
				TenantModel: models.TenantModel{TenantID: 0}, // 系统角色
				Code:        roleData.Code,
				Name:        roleData.Name,
				Description: roleData.Description,
				Type:        roleData.Type,
				IsActive:    true,
			}

			if err := db.Create(&role).Error; err != nil {
				return fmt.Errorf("创建角色失败 %s: %v", roleData.Code, err)
			}

			// 分配权限给角色
			if err := assignPermissionsToRole(db, role.ID, roleData.Permissions); err != nil {
				return fmt.Errorf("分配权限失败 %s: %v", roleData.Code, err)
			}

			log.Printf("创建系统角色: %s - %s", roleData.Code, roleData.Name)
		} else if result.Error != nil {
			return fmt.Errorf("查询角色失败 %s: %v", roleData.Code, result.Error)
		} else {
			// 角色已存在，更新基本信息和权限
			updates := map[string]interface{}{
				"name":        roleData.Name,
				"description": roleData.Description,
				"type":        roleData.Type,
			}

			if err := db.Model(&existingRole).Updates(updates).Error; err != nil {
				return fmt.Errorf("更新角色失败 %s: %v", roleData.Code, err)
			}

			// 重新分配权限
			if err := clearRolePermissions(db, existingRole.ID); err != nil {
				return fmt.Errorf("清除角色权限失败 %s: %v", roleData.Code, err)
			}

			if err := assignPermissionsToRole(db, existingRole.ID, roleData.Permissions); err != nil {
				return fmt.Errorf("重新分配权限失败 %s: %v", roleData.Code, err)
			}

			log.Printf("更新系统角色: %s - %s", roleData.Code, roleData.Name)
		}
	}

	return nil
}

// assignPermissionsToRole 分配权限给角色
func assignPermissionsToRole(db *gorm.DB, roleID uint64, permissionCodes []string) error {
	for _, code := range permissionCodes {
		var permission models.Permission
		if err := db.Where("code = ?", code).First(&permission).Error; err != nil {
			return fmt.Errorf("找不到权限 %s: %v", code, err)
		}

		rolePerm := models.RolePermission{
			RoleID:       roleID,
			PermissionID: permission.ID,
		}

		if err := db.Create(&rolePerm).Error; err != nil {
			return fmt.Errorf("分配权限失败 %s: %v", code, err)
		}
	}

	return nil
}

// clearRolePermissions 清除角色权限
func clearRolePermissions(db *gorm.DB, roleID uint64) error {
	return db.Where("role_id = ?", roleID).Delete(&models.RolePermission{}).Error
}
