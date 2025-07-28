// Package main provides database migration tools for the Shield project.
// This file specifically handles the initialization of system permissions and roles.
//
// Permission Hierarchy:
// - Menu permissions: Top-level navigation access
// - Button permissions: UI component access (child of menu)
// - API permissions: Backend endpoint access (child of button)
// - Field permissions: Data field visibility control
//
// Permission Scopes:
// - System scope: Platform-level permissions (tenant_id = 0)
// - Tenant scope: Organization-level permissions (tenant_id > 0)
//
// The dynamic permission validation system uses this data to:
// 1. Check API access based on resource_path and method matching
// 2. Grant automatic access to system tenant users (tenant_id = 0)
// 3. Require explicit database permissions for regular tenant users
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
	// 菜单专用字段
	MenuIcon      string
	MenuPath      string
	MenuComponent string
	MenuVisible   bool
}

// InitSystemPermissions 初始化系统权限
func InitSystemPermissions(db *gorm.DB) error {
	permissions := getSystemPermissions()
	
	// 验证权限数据一致性
	if err := ValidatePermissionData(permissions); err != nil {
		return fmt.Errorf("权限数据验证失败: %v", err)
	}

	for _, permData := range permissions {
		var existingPerm models.Permission
		result := db.Where("code = ?", permData.Code).First(&existingPerm)

		if result.Error == gorm.ErrRecordNotFound {
			// 权限不存在，创建新权限
			// 为非菜单权限设置默认值
			menuVisible := permData.MenuVisible
			if permData.Type != models.PermissionTypeMenu {
				menuVisible = false
			}

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
				// 菜单专用字段
				MenuIcon:      permData.MenuIcon,
				MenuPath:      permData.MenuPath,
				MenuComponent: permData.MenuComponent,
				MenuVisible:   menuVisible,
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
				// 菜单专用字段
				"menu_icon":      permData.MenuIcon,
				"menu_path":      permData.MenuPath,
				"menu_component": permData.MenuComponent,
				"menu_visible":   permData.MenuVisible,
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
// 
// 权限组织结构说明：
// 1. 系统管理权限 (1000-1999): 平台级管理功能
//    - 租户管理 (1010-1099): 租户CRUD操作
//    - 权限管理 (1020-1099): 权限配置管理
//    - 系统监控 (1030-1099): 健康检查和状态监控
// 
// 2. 认证授权权限 (1900-1999): 身份验证和授权相关
//    - 登录登出 (1900-1910): 基础认证流程
//    - 验证码 (1903-1910): 安全验证机制
// 
// 3. 租户业务权限 (2000+): 租户内业务功能
//    - 用户管理 (2000-2999): 用户CRUD和资料管理
//    - 角色管理 (3000-3999): 角色和权限分配
func getSystemPermissions() []PermissionData {
	return []PermissionData{
		// 系统管理模块
		{
			Code:         "system_menu",
			Name:         "系统管理",
			Description:  "系统管理菜单访问权限",
			Type:         models.PermissionTypeMenu,
			Scope:        models.ScopeSystem,
			ParentCode:   "",
			SortOrder:    1000,
			Module:       models.ModuleSystem,
			MenuIcon:     "Settings",
			MenuPath:     "/system",
			MenuComponent: "SystemLayout",
			MenuVisible:  true,
		},

		// 租户管理
		{
			Code:         "tenant_menu",
			Name:         "租户管理",
			Description:  "租户管理菜单访问权限",
			Type:         models.PermissionTypeMenu,
			Scope:        models.ScopeSystem,
			ParentCode:   "system_menu",
			SortOrder:    1010,
			Module:       models.ModuleTenant,
			MenuIcon:     "Building",
			MenuPath:     "/system/tenants",
			MenuComponent: "TenantManagement",
			MenuVisible:  true,
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
			Code:         "permission_menu",
			Name:         "权限管理",
			Description:  "权限管理菜单访问权限",
			Type:         models.PermissionTypeMenu,
			Scope:        models.ScopeSystem,
			ParentCode:   "system_menu",
			SortOrder:    1020,
			Module:       models.ModuleSystem,
			MenuIcon:     "Shield",
			MenuPath:     "/system/permissions",
			MenuComponent: "PermissionManagement",
			MenuVisible:  true,
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

		// 系统监控和健康检查
		{
			Code:         "health_check_api",
			Name:         "健康检查API",
			Description:  "系统健康检查API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeSystem,
			ParentCode:   "system_menu",
			ResourcePath: "/health",
			Method:       "GET",
			SortOrder:    1030,
			Module:       models.ModuleSystem,
		},
		{
			Code:         "system_status_api",
			Name:         "系统状态API",
			Description:  "获取系统状态API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeSystem,
			ParentCode:   "system_menu",
			ResourcePath: "/api/v1/system/status",
			Method:       "GET",
			SortOrder:    1031,
			Module:       models.ModuleSystem,
		},

		// 认证模块权限
		{
			Code:         "auth_login_api",
			Name:         "用户登录API",
			Description:  "用户登录API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "",
			ResourcePath: "/api/v1/auth/login",
			Method:       "POST",
			SortOrder:    1900,
			Module:       models.ModuleAuth,
		},
		{
			Code:         "auth_logout_api",
			Name:         "用户登出API",
			Description:  "用户登出API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "",
			ResourcePath: "/api/v1/auth/logout",
			Method:       "POST",
			SortOrder:    1901,
			Module:       models.ModuleAuth,
		},
		{
			Code:         "auth_refresh_api",
			Name:         "刷新Token API",
			Description:  "刷新访问令牌API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "",
			ResourcePath: "/api/v1/auth/refresh",
			Method:       "POST",
			SortOrder:    1902,
			Module:       models.ModuleAuth,
		},
		{
			Code:         "captcha_generate_api",
			Name:         "验证码生成API",
			Description:  "生成验证码API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "",
			ResourcePath: "/api/v1/captcha/generate",
			Method:       "GET",
			SortOrder:    1903,
			Module:       models.ModuleAuth,
		},
		{
			Code:         "captcha_verify_api",
			Name:         "验证码验证API",
			Description:  "验证验证码API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "",
			ResourcePath: "/api/v1/captcha/verify",
			Method:       "POST",
			SortOrder:    1904,
			Module:       models.ModuleAuth,
		},

		// 租户权限
		// 用户管理模块
		{
			Code:         "user_menu",
			Name:         "用户管理",
			Description:  "用户管理菜单访问权限",
			Type:         models.PermissionTypeMenu,
			Scope:        models.ScopeTenant,
			ParentCode:   "",
			SortOrder:    2000,
			Module:       models.ModuleUser,
			MenuIcon:     "Users",
			MenuPath:     "/users",
			MenuComponent: "UserManagement",
			MenuVisible:  true,
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
		{
			Code:        "user_profile_btn",
			Name:        "个人资料",
			Description: "查看个人资料按钮权限",
			Type:        models.PermissionTypeButton,
			Scope:       models.ScopeTenant,
			ParentCode:  "user_menu",
			SortOrder:   2050,
			Module:      models.ModuleUser,
		},
		{
			Code:         "user_profile_api",
			Name:         "个人资料API",
			Description:  "获取个人资料API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "user_profile_btn",
			ResourcePath: "/api/v1/users/profile",
			Method:       "GET",
			SortOrder:    2051,
			Module:       models.ModuleUser,
		},
		{
			Code:         "user_profile_update_api",
			Name:         "更新个人资料API",
			Description:  "更新个人资料API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "user_profile_btn",
			ResourcePath: "/api/v1/users/profile",
			Method:       "PUT",
			SortOrder:    2052,
			Module:       models.ModuleUser,
		},
		{
			Code:         "user_password_change_api",
			Name:         "修改密码API",
			Description:  "修改用户密码API权限",
			Type:         models.PermissionTypeAPI,
			Scope:        models.ScopeTenant,
			ParentCode:   "user_profile_btn",
			ResourcePath: "/api/v1/users/change-password",
			Method:       "POST",
			SortOrder:    2053,
			Module:       models.ModuleUser,
		},

		// 角色管理模块
		{
			Code:         "role_menu",
			Name:         "角色管理",
			Description:  "角色管理菜单访问权限",
			Type:         models.PermissionTypeMenu,
			Scope:        models.ScopeTenant,
			ParentCode:   "",
			SortOrder:    3000,
			Module:       models.ModuleRole,
			MenuIcon:     "UserCheck",
			MenuPath:     "/roles",
			MenuComponent: "RoleManagement",
			MenuVisible:  true,
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
	permissions := getSystemPermissions()
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
				"permission_update_api", "health_check_api", "system_status_api",
			},
		},
		{
			Code:        models.RoleTenantAdmin,
			Name:        "租户管理员",
			Description: "租户管理员，拥有租户内所有权限",
			Type:        models.RoleTypeSystem,
			Permissions: []string{
				// 认证权限
				"auth_login_api", "auth_logout_api", "auth_refresh_api",
				"captcha_generate_api", "captcha_verify_api",
				// 用户管理权限
				"user_menu", "user_list_btn", "user_list_api", "user_create_btn", "user_create_api",
				"user_update_btn", "user_update_api", "user_delete_btn", "user_delete_api",
				"user_profile_btn", "user_profile_api", "user_profile_update_api", "user_password_change_api",
				// 角色管理权限
				"role_menu", "role_list_btn", "role_list_api", "role_create_btn", "role_create_api",
				"role_assign_btn", "role_assign_api", "field_permission_btn", "field_permission_list_api",
				"field_permission_update_api",
			},
		},
	}
	
	// 验证角色权限配置
	if err := ValidateRolePermissions(roles, permissions); err != nil {
		return fmt.Errorf("角色权限配置验证失败: %v", err)
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

// ValidatePermissionData 验证权限数据的一致性
func ValidatePermissionData(permissions []PermissionData) error {
	codes := make(map[string]bool)
	parentCodes := make(map[string]bool)
	
	// 收集所有权限代码
	for _, perm := range permissions {
		if codes[perm.Code] {
			return fmt.Errorf("发现重复的权限代码: %s", perm.Code)
		}
		codes[perm.Code] = true
		
		if perm.ParentCode != "" {
			parentCodes[perm.ParentCode] = true
		}
	}
	
	// 检查父权限是否存在
	for parentCode := range parentCodes {
		if !codes[parentCode] {
			return fmt.Errorf("找不到父权限: %s", parentCode)
		}
	}
	
	// 验证API权限必须有resource_path和method
	for _, perm := range permissions {
		if perm.Type == models.PermissionTypeAPI {
			if perm.ResourcePath == "" {
				return fmt.Errorf("API权限 %s 缺少resource_path", perm.Code)
			}
			if perm.Method == "" {
				return fmt.Errorf("API权限 %s 缺少method", perm.Code)
			}
		}
	}
	
	log.Printf("权限数据验证通过，共 %d 个权限", len(permissions))
	return nil
}

// ValidateRolePermissions 验证角色权限配置
func ValidateRolePermissions(roles []struct {
	Code        string
	Name        string
	Description string
	Type        string
	Permissions []string
}, permissions []PermissionData) error {
	// 建立权限代码映射
	permissionCodes := make(map[string]bool)
	for _, perm := range permissions {
		permissionCodes[perm.Code] = true
	}
	
	// 验证角色权限是否存在
	for _, role := range roles {
		for _, permCode := range role.Permissions {
			if !permissionCodes[permCode] {
				return fmt.Errorf("角色 %s 引用了不存在的权限: %s", role.Code, permCode)
			}
		}
		log.Printf("角色 %s 权限验证通过，共 %d 个权限", role.Code, len(role.Permissions))
	}
	
	return nil
}
