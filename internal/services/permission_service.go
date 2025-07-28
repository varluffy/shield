// Package services contains business logic and service layer implementations.
// It provides the core application functionality and business rules.
package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/internal/repositories"
	"github.com/varluffy/shield/pkg/logger"
	"go.uber.org/zap"
)

// PermissionService 权限服务接口
type PermissionService interface {
	// CheckUserPermission 检查用户是否拥有指定权限
	CheckUserPermission(ctx context.Context, userID, tenantID, permissionCode string) (bool, error)
	// CheckUserAPIPermission 检查用户是否有访问指定API的权限
	CheckUserAPIPermission(ctx context.Context, userID, tenantID, path, method string) (bool, error)
	// GetUserRoles 获取用户角色
	GetUserRoles(ctx context.Context, userID, tenantID string) ([]models.Role, error)
	// GetUserPermissions 获取用户权限
	GetUserPermissions(ctx context.Context, userID, tenantID string) ([]models.Permission, error)
	// HasRole 检查用户是否拥有指定角色
	HasRole(ctx context.Context, userID, tenantID, roleCode string) (bool, error)
	// IsSystemAdmin 检查用户是否为系统管理员
	IsSystemAdmin(ctx context.Context, userID string) (bool, error)
	// IsTenantAdmin 检查用户是否为租户管理员
	IsTenantAdmin(ctx context.Context, userID, tenantID string) (bool, error)

	// 权限管理方法
	// ListPermissions 获取权限列表
	ListPermissions(ctx context.Context, filter map[string]interface{}, page, limit int) ([]models.Permission, int64, error)
	// GetPermissionTree 获取权限树结构
	GetPermissionTree(ctx context.Context, tenantID, scope string) (interface{}, error)
	// ListSystemPermissions 获取系统权限列表
	ListSystemPermissions(ctx context.Context, page, limit int) ([]models.Permission, int64, error)
	// UpdatePermission 更新权限
	UpdatePermission(ctx context.Context, id uint64, name, description string, isActive *bool) (*models.Permission, error)
	// CreatePermission 创建权限
	CreatePermission(ctx context.Context, permission *models.Permission) error
	// DeletePermission 删除权限
	DeletePermission(ctx context.Context, id uint64) error
	// GetPermissionByCode 根据代码获取权限
	GetPermissionByCode(ctx context.Context, code string) (*models.Permission, error)
	// GetPermissionByID 根据ID获取权限
	GetPermissionByID(ctx context.Context, id uint64) (*models.Permission, error)
	
	// 菜单权限方法
	// GetUserMenuPermissions 获取用户菜单权限
	GetUserMenuPermissions(ctx context.Context, userID, tenantID string) ([]models.Permission, error)
	// BuildMenuTree 构建菜单树结构
	BuildMenuTree(ctx context.Context, permissions []models.Permission) []interface{}
	// GetMenuPermissionsByScope 根据范围获取菜单权限
	GetMenuPermissionsByScope(ctx context.Context, scope string) ([]models.Permission, error)
}

// permissionService 权限服务实现
type permissionService struct {
	userRepo       repositories.UserRepository
	roleRepo       repositories.RoleRepository
	permissionRepo repositories.PermissionRepository
	tenantRepo     repositories.TenantRepository
	cacheService   PermissionCacheService
	logger         *logger.Logger
}

// NewPermissionService 创建权限服务
func NewPermissionService(
	userRepo repositories.UserRepository,
	roleRepo repositories.RoleRepository,
	permissionRepo repositories.PermissionRepository,
	tenantRepo repositories.TenantRepository,
	cacheService PermissionCacheService,
	logger *logger.Logger,
) PermissionService {
	return &permissionService{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		tenantRepo:     tenantRepo,
		cacheService:   cacheService,
		logger:         logger,
	}
}

// CheckUserPermission 检查用户是否拥有指定权限
func (s *permissionService) CheckUserPermission(ctx context.Context, userID, tenantID, permissionCode string) (bool, error) {
	s.logger.DebugWithTrace(ctx, "Checking user permission",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.String("permission_code", permissionCode))

	// 获取用户权限
	permissions, err := s.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get user permissions",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return false, err
	}

	// 检查是否拥有指定权限
	for _, permission := range permissions {
		if permission.Code == permissionCode {
			s.logger.DebugWithTrace(ctx, "Permission granted",
				zap.String("user_id", userID),
				zap.String("permission_code", permissionCode))
			return true, nil
		}
	}

	s.logger.DebugWithTrace(ctx, "Permission denied",
		zap.String("user_id", userID),
		zap.String("permission_code", permissionCode))
	return false, nil
}

// GetUserRoles 获取用户角色
func (s *permissionService) GetUserRoles(ctx context.Context, userID, tenantID string) ([]models.Role, error) {
	s.logger.DebugWithTrace(ctx, "Getting user roles",
		zap.String("user_uuid", userID),
		zap.String("tenant_id", tenantID))

	// 尝试从缓存获取
	if s.cacheService != nil {
		cachedRoles, err := s.cacheService.GetUserRoles(ctx, userID, tenantID)
		if err == nil && cachedRoles != nil {
			s.logger.DebugWithTrace(ctx, "Retrieved user roles from cache",
				zap.String("user_uuid", userID),
				zap.Int("role_count", len(cachedRoles)))
			return cachedRoles, nil
		}
	}

	// 转换tenantID为UUID（如果传入的是数字ID）
	tenantUUID, err := s.convertTenantIDToUUID(ctx, tenantID)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to convert tenant ID",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to convert tenant ID: %w", err)
	}

	roles, err := s.roleRepo.GetUserRolesByUUID(ctx, userID, tenantUUID)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get user roles",
			zap.String("user_uuid", userID),
			zap.String("tenant_uuid", tenantUUID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	// 缓存结果
	if s.cacheService != nil {
		_ = s.cacheService.SetUserRoles(ctx, userID, tenantID, roles)
	}

	s.logger.DebugWithTrace(ctx, "Retrieved user roles",
		zap.String("user_uuid", userID),
		zap.Int("role_count", len(roles)))

	return roles, nil
}

// GetUserPermissions 获取用户权限
func (s *permissionService) GetUserPermissions(ctx context.Context, userID, tenantID string) ([]models.Permission, error) {
	s.logger.DebugWithTrace(ctx, "Getting user permissions",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID))

	// 尝试从缓存获取
	if s.cacheService != nil {
		cachedPermissions, err := s.cacheService.GetUserPermissions(ctx, userID, tenantID)
		if err == nil && cachedPermissions != nil {
			s.logger.DebugWithTrace(ctx, "Retrieved user permissions from cache",
				zap.String("user_id", userID),
				zap.Int("permission_count", len(cachedPermissions)))
			return cachedPermissions, nil
		}
	}

	// 转换tenantID为UUID（如果传入的是数字ID）
	tenantUUID, err := s.convertTenantIDToUUID(ctx, tenantID)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to convert tenant ID",
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to convert tenant ID: %w", err)
	}

	permissions, err := s.permissionRepo.GetUserPermissions(ctx, userID, tenantUUID)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get user permissions",
			zap.String("user_id", userID),
			zap.String("tenant_uuid", tenantUUID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// 缓存结果
	if s.cacheService != nil {
		_ = s.cacheService.SetUserPermissions(ctx, userID, tenantID, permissions)
	}

	s.logger.DebugWithTrace(ctx, "Retrieved user permissions",
		zap.String("user_id", userID),
		zap.Int("permission_count", len(permissions)))

	return permissions, nil
}

// HasRole 检查用户是否拥有指定角色
func (s *permissionService) HasRole(ctx context.Context, userID, tenantID, roleCode string) (bool, error) {
	s.logger.DebugWithTrace(ctx, "Checking user role",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.String("role_code", roleCode))

	roles, err := s.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Code == roleCode {
			s.logger.DebugWithTrace(ctx, "Role found",
				zap.String("user_id", userID),
				zap.String("role_code", roleCode))
			return true, nil
		}
	}

	s.logger.DebugWithTrace(ctx, "Role not found",
		zap.String("user_id", userID),
		zap.String("role_code", roleCode))
	return false, nil
}

// IsSystemAdmin 检查用户是否为系统管理员
func (s *permissionService) IsSystemAdmin(ctx context.Context, userID string) (bool, error) {
	s.logger.DebugWithTrace(ctx, "Checking if user is system admin",
		zap.String("user_uuid", userID))

	// 获取用户信息
	user, err := s.userRepo.GetByUUID(ctx, userID)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get user",
			zap.String("user_uuid", userID),
			zap.Error(err))
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查用户是否在租户ID为0的情况下拥有系统管理员角色
	// 系统管理员应该属于tenant_id=0（系统租户）
	if user.TenantID != 0 {
		s.logger.DebugWithTrace(ctx, "User is not in system tenant",
			zap.String("user_uuid", userID),
			zap.Uint64("tenant_id", user.TenantID))
		return false, nil
	}

	// 检查系统租户中的系统管理员角色
	roles, err := s.roleRepo.GetUserRoles(ctx, user.ID, 0) // tenant_id=0 为系统租户
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get user roles in system tenant",
			zap.String("user_uuid", userID),
			zap.Error(err))
		return false, fmt.Errorf("failed to get user roles: %w", err)
	}

	// 检查是否有系统管理员角色
	for _, role := range roles {
		if role.Code == "system_admin" && role.Type == "system" {
			s.logger.DebugWithTrace(ctx, "User is system admin",
				zap.String("user_uuid", userID),
				zap.String("role_code", role.Code))
			return true, nil
		}
	}

	s.logger.DebugWithTrace(ctx, "User is not system admin",
		zap.String("user_uuid", userID))
	return false, nil
}

// IsTenantAdmin 检查用户是否为租户管理员
func (s *permissionService) IsTenantAdmin(ctx context.Context, userID, tenantID string) (bool, error) {
	s.logger.DebugWithTrace(ctx, "Checking if user is tenant admin",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID))

	// 检查是否为系统管理员（系统管理员拥有所有租户的管理权限）
	isSystemAdmin, err := s.IsSystemAdmin(ctx, userID)
	if err != nil {
		return false, err
	}
	if isSystemAdmin {
		return true, nil
	}

	// 检查是否为指定租户的管理员
	return s.HasRole(ctx, userID, tenantID, "tenant_admin")
}

// convertTenantIDToUUID 将租户ID转换为UUID
// 如果传入的已经是UUID格式，直接返回；如果是数字ID，查询获取UUID
func (s *permissionService) convertTenantIDToUUID(ctx context.Context, tenantID string) (string, error) {
	// 检查是否已经是UUID格式（36位字符串包含短横线）
	if len(tenantID) == 36 && tenantID[8] == '-' && tenantID[13] == '-' {
		return tenantID, nil
	}

	// 否则，假设是数字ID，需要查询获取UUID
	numericID, err := strconv.ParseUint(tenantID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid tenant ID format: %s", tenantID)
	}

	// 特殊处理系统级别（ID为0）
	if numericID == 0 {
		// 系统级别，返回系统标识符
		return "00000000-0000-0000-0000-000000000000", nil
	}

	// 通过TenantRepository查询租户UUID
	tenantUUID, err := s.tenantRepo.GetUUIDByID(ctx, numericID)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get tenant UUID by ID",
			zap.Uint64("tenant_id", numericID),
			zap.Error(err))
		return "", fmt.Errorf("failed to get tenant UUID for ID %d: %w", numericID, err)
	}

	s.logger.DebugWithTrace(ctx, "Converted tenant ID to UUID",
		zap.Uint64("tenant_id", numericID),
		zap.String("tenant_uuid", tenantUUID))

	return tenantUUID, nil
}

// ListPermissions 获取权限列表（根据用户身份自动过滤）
func (s *permissionService) ListPermissions(ctx context.Context, filter map[string]interface{}, page, limit int) ([]models.Permission, int64, error) {
	s.logger.DebugWithTrace(ctx, "Listing permissions",
		zap.Any("filter", filter),
		zap.Int("page", page),
		zap.Int("limit", limit))

	// 从上下文获取用户信息
	userID := ctx.Value("user_id")
	if userID != nil {
		userIDStr, ok := userID.(string)
		if ok {
			// 检查是否为系统管理员
			isSystemAdmin, err := s.IsSystemAdmin(ctx, userIDStr)
			if err != nil {
				s.logger.WarnWithTrace(ctx, "Failed to check if user is system admin",
					zap.String("user_id", userIDStr),
					zap.Error(err))
			} else if !isSystemAdmin {
				// 如果不是系统管理员，只返回租户权限
				filter["scope"] = "tenant"
				s.logger.DebugWithTrace(ctx, "User is not system admin, filtering tenant permissions only",
					zap.String("user_id", userIDStr))
			} else {
				s.logger.DebugWithTrace(ctx, "User is system admin, returning all permissions",
					zap.String("user_id", userIDStr))
			}
		}
	}

	permissions, total, err := s.permissionRepo.List(ctx, filter, page, limit)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to list permissions",
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list permissions: %w", err)
	}

	s.logger.DebugWithTrace(ctx, "Retrieved permissions",
		zap.Int("count", len(permissions)),
		zap.Int64("total", total))

	return permissions, total, nil
}

// GetPermissionTree 获取权限树结构（根据用户身份自动过滤）
func (s *permissionService) GetPermissionTree(ctx context.Context, tenantID, scope string) (interface{}, error) {
	s.logger.DebugWithTrace(ctx, "Getting permission tree",
		zap.String("tenant_id", tenantID),
		zap.String("scope", scope))

	// 从上下文获取用户信息，自动判断scope
	userID := ctx.Value("user_id")
	finalScope := scope
	if userID != nil {
		userIDStr, ok := userID.(string)
		if ok {
			// 检查是否为系统管理员
			isSystemAdmin, err := s.IsSystemAdmin(ctx, userIDStr)
			if err != nil {
				s.logger.WarnWithTrace(ctx, "Failed to check if user is system admin",
					zap.String("user_id", userIDStr),
					zap.Error(err))
			} else if !isSystemAdmin {
				// 如果不是系统管理员，强制使用tenant scope
				finalScope = "tenant"
				s.logger.DebugWithTrace(ctx, "User is not system admin, using tenant scope",
					zap.String("user_id", userIDStr))
			} else {
				// 系统管理员：如果没有指定scope，返回所有权限
				if scope == "" {
					finalScope = "" // 返回所有权限
				}
				s.logger.DebugWithTrace(ctx, "User is system admin, using provided scope",
					zap.String("user_id", userIDStr),
					zap.String("final_scope", finalScope))
			}
		}
	}

	// 构建查询条件
	filter := map[string]interface{}{}
	if finalScope != "" {
		filter["scope"] = finalScope
	}

	permissions, _, err := s.permissionRepo.List(ctx, filter, 1, 1000) // 获取所有权限
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get permissions for tree",
			zap.Error(err))
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	// 构建权限树
	tree := s.buildPermissionTree(permissions)

	s.logger.DebugWithTrace(ctx, "Built permission tree",
		zap.String("final_scope", finalScope),
		zap.Int("total_permissions", len(permissions)))

	return tree, nil
}

// ListSystemPermissions 获取系统权限列表
func (s *permissionService) ListSystemPermissions(ctx context.Context, page, limit int) ([]models.Permission, int64, error) {
	s.logger.DebugWithTrace(ctx, "Listing system permissions",
		zap.Int("page", page),
		zap.Int("limit", limit))

	filter := map[string]interface{}{
		"scope": "system", // 只获取系统权限
	}

	permissions, total, err := s.permissionRepo.List(ctx, filter, page, limit)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to list system permissions",
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list system permissions: %w", err)
	}

	s.logger.DebugWithTrace(ctx, "Retrieved system permissions",
		zap.Int("count", len(permissions)),
		zap.Int64("total", total))

	return permissions, total, nil
}

// UpdatePermission 更新权限
func (s *permissionService) UpdatePermission(ctx context.Context, id uint64, name, description string, isActive *bool) (*models.Permission, error) {
	s.logger.DebugWithTrace(ctx, "Updating permission",
		zap.Uint64("id", id),
		zap.String("name", name),
		zap.String("description", description))

	// 获取现有权限
	permission, err := s.permissionRepo.GetByNumericID(ctx, id)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get permission for update",
			zap.Uint64("id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	// 更新字段
	permission.Name = name
	permission.Description = description
	if isActive != nil {
		permission.IsActive = *isActive
	}

	// 保存更新
	updatedPermission, err := s.permissionRepo.Update(ctx, permission)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to update permission",
			zap.Uint64("id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	s.logger.InfoWithTrace(ctx, "Permission updated successfully",
		zap.Uint64("id", id),
		zap.String("name", name))

	return updatedPermission, nil
}

// buildPermissionTree 构建权限树结构
func (s *permissionService) buildPermissionTree(permissions []models.Permission) []map[string]interface{} {
	// 创建权限映射
	permMap := make(map[string]models.Permission)
	for _, perm := range permissions {
		permMap[perm.Code] = perm
	}

	// 找到根节点（没有父权限的节点）
	var roots []map[string]interface{}
	processed := make(map[string]bool)

	for _, perm := range permissions {
		if perm.ParentCode == "" && !processed[perm.Code] {
			node := s.buildTreeNode(perm, permMap, processed)
			roots = append(roots, node)
		}
	}

	return roots
}

// buildTreeNode 构建树节点
func (s *permissionService) buildTreeNode(perm models.Permission, permMap map[string]models.Permission, processed map[string]bool) map[string]interface{} {
	processed[perm.Code] = true

	node := map[string]interface{}{
		"id":          perm.ID,
		"code":        perm.Code,
		"name":        perm.Name,
		"description": perm.Description,
		"type":        perm.Type,
		"scope":       perm.Scope,
		"module":      perm.Module,
		"sort_order":  perm.SortOrder,
		"is_active":   perm.IsActive,
		"children":    []map[string]interface{}{},
	}

	// 查找子节点
	var children []map[string]interface{}
	for _, p := range permMap {
		if p.ParentCode == perm.Code && !processed[p.Code] {
			childNode := s.buildTreeNode(p, permMap, processed)
			children = append(children, childNode)
		}
	}

	node["children"] = children
	return node
}

// CreatePermission 创建权限
func (s *permissionService) CreatePermission(ctx context.Context, permission *models.Permission) error {
	s.logger.DebugWithTrace(ctx, "Creating permission",
		zap.String("code", permission.Code),
		zap.String("name", permission.Name))

	// 检查权限代码是否已存在
	existing, err := s.permissionRepo.GetByCode(ctx, permission.Code)
	if err == nil && existing != nil {
		s.logger.WarnWithTrace(ctx, "Permission code already exists",
			zap.String("code", permission.Code))
		return fmt.Errorf("permission code already exists: %s", permission.Code)
	}

	err = s.permissionRepo.Create(ctx, permission)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to create permission",
			zap.String("code", permission.Code),
			zap.Error(err))
		return fmt.Errorf("failed to create permission: %w", err)
	}

	s.logger.InfoWithTrace(ctx, "Permission created successfully",
		zap.String("code", permission.Code),
		zap.String("name", permission.Name))

	return nil
}

// DeletePermission 删除权限
func (s *permissionService) DeletePermission(ctx context.Context, id uint64) error {
	s.logger.DebugWithTrace(ctx, "Deleting permission",
		zap.Uint64("id", id))

	// 检查权限是否存在
	permission, err := s.permissionRepo.GetByNumericID(ctx, id)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get permission for deletion",
			zap.Uint64("id", id),
			zap.Error(err))
		return fmt.Errorf("failed to get permission: %w", err)
	}

	// 检查是否为内置权限
	if permission.IsBuiltin {
		s.logger.WarnWithTrace(ctx, "Cannot delete builtin permission",
			zap.Uint64("id", id),
			zap.String("code", permission.Code))
		return fmt.Errorf("cannot delete builtin permission: %s", permission.Code)
	}

	err = s.permissionRepo.Delete(ctx, fmt.Sprintf("%d", id))
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to delete permission",
			zap.Uint64("id", id),
			zap.Error(err))
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	s.logger.InfoWithTrace(ctx, "Permission deleted successfully",
		zap.Uint64("id", id),
		zap.String("code", permission.Code))

	return nil
}

// GetPermissionByCode 根据代码获取权限
func (s *permissionService) GetPermissionByCode(ctx context.Context, code string) (*models.Permission, error) {
	s.logger.DebugWithTrace(ctx, "Getting permission by code",
		zap.String("code", code))

	permission, err := s.permissionRepo.GetByCode(ctx, code)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get permission by code",
			zap.String("code", code),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	s.logger.DebugWithTrace(ctx, "Retrieved permission by code",
		zap.String("code", code),
		zap.String("name", permission.Name))

	return permission, nil
}

// CheckUserAPIPermission 检查用户是否有访问指定API的权限
func (s *permissionService) CheckUserAPIPermission(ctx context.Context, userID, tenantID, path, method string) (bool, error) {
	s.logger.DebugWithTrace(ctx, "Checking user API permission",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.String("path", path),
		zap.String("method", method))

	// 1. 获取匹配该路径和方法的权限
	permissions, err := s.permissionRepo.GetPermissionsByPathAndMethod(ctx, path, method)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get permissions by path and method",
			zap.String("path", path),
			zap.String("method", method),
			zap.Error(err))
		return false, fmt.Errorf("failed to get permissions: %w", err)
	}

	// 如果没有找到匹配的权限，默认拒绝访问
	if len(permissions) == 0 {
		s.logger.DebugWithTrace(ctx, "No permissions found for API path",
			zap.String("path", path),
			zap.String("method", method))
		return false, nil
	}

	// 2. 获取用户的所有权限
	userPermissions, err := s.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get user permissions",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return false, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// 3. 创建用户权限代码映射，用于快速查找
	userPermissionCodes := make(map[string]bool)
	for _, perm := range userPermissions {
		userPermissionCodes[perm.Code] = true
	}

	// 4. 检查用户是否拥有任意一个匹配的权限
	for _, apiPermission := range permissions {
		if userPermissionCodes[apiPermission.Code] {
			s.logger.DebugWithTrace(ctx, "API permission granted",
				zap.String("user_id", userID),
				zap.String("path", path),
				zap.String("method", method),
				zap.String("permission_code", apiPermission.Code))
			return true, nil
		}
	}

	// 记录所有匹配的权限代码，用于调试
	var requiredPermissions []string
	for _, perm := range permissions {
		requiredPermissions = append(requiredPermissions, perm.Code)
	}

	s.logger.DebugWithTrace(ctx, "API permission denied",
		zap.String("user_id", userID),
		zap.String("path", path),
		zap.String("method", method),
		zap.Strings("required_permissions", requiredPermissions))

	return false, nil
}

// GetPermissionByID 根据ID获取权限
func (s *permissionService) GetPermissionByID(ctx context.Context, id uint64) (*models.Permission, error) {
	s.logger.DebugWithTrace(ctx, "Getting permission by ID",
		zap.Uint64("id", id))

	permission, err := s.permissionRepo.GetByNumericID(ctx, id)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get permission by ID",
			zap.Uint64("id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	s.logger.DebugWithTrace(ctx, "Retrieved permission by ID",
		zap.Uint64("id", id),
		zap.String("code", permission.Code))

	return permission, nil
}

// GetUserMenuPermissions 获取用户菜单权限
func (s *permissionService) GetUserMenuPermissions(ctx context.Context, userID, tenantID string) ([]models.Permission, error) {
	s.logger.DebugWithTrace(ctx, "Getting user menu permissions",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID))

	// 检查是否为系统管理员
	isSystemAdmin, err := s.IsSystemAdmin(ctx, userID)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to check if user is system admin",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to check system admin status: %w", err)
	}

	var menuPermissions []models.Permission

	if isSystemAdmin {
		// 系统管理员：获取系统菜单权限
		systemMenus, err := s.GetMenuPermissionsByScope(ctx, "system")
		if err != nil {
			s.logger.ErrorWithTrace(ctx, "Failed to get system menu permissions",
				zap.String("user_id", userID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to get system menu permissions: %w", err)
		}
		menuPermissions = append(menuPermissions, systemMenus...)
	}

	// 获取用户在租户内的菜单权限
	userPermissions, err := s.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get user permissions",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	// 过滤出菜单类型的权限
	for _, perm := range userPermissions {
		if perm.Type == "menu" {
			menuPermissions = append(menuPermissions, perm)
		}
	}

	s.logger.DebugWithTrace(ctx, "Retrieved user menu permissions",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.Int("menu_count", len(menuPermissions)),
		zap.Bool("is_system_admin", isSystemAdmin))

	return menuPermissions, nil
}

// BuildMenuTree 构建菜单树结构
func (s *permissionService) BuildMenuTree(ctx context.Context, permissions []models.Permission) []interface{} {
	s.logger.DebugWithTrace(ctx, "Building menu tree",
		zap.Int("permission_count", len(permissions)))

	// 创建权限映射
	permMap := make(map[string]models.Permission)
	for _, perm := range permissions {
		permMap[perm.Code] = perm
	}

	// 找到根节点（没有父权限的节点）
	var roots []interface{}
	processed := make(map[string]bool)

	for _, perm := range permissions {
		if perm.ParentCode == "" && !processed[perm.Code] {
			node := s.buildMenuTreeNode(perm, permMap, processed)
			roots = append(roots, node)
		}
	}

	s.logger.DebugWithTrace(ctx, "Built menu tree",
		zap.Int("root_count", len(roots)))

	return roots
}

// buildMenuTreeNode 构建菜单树节点
func (s *permissionService) buildMenuTreeNode(perm models.Permission, permMap map[string]models.Permission, processed map[string]bool) map[string]interface{} {
	processed[perm.Code] = true

	// 优先使用数据库中的菜单配置，否则使用生成策略
	path := perm.MenuPath
	if path == "" {
		path = s.generateMenuPath(perm)
	}
	
	icon := perm.MenuIcon  
	if icon == "" {
		icon = s.generateMenuIcon(perm)
	}
	
	// 调试输出（已完成）
	s.logger.DebugWithTrace(context.Background(), "Menu tree node data",
		zap.String("code", perm.Code),
		zap.String("db_icon", perm.MenuIcon),
		zap.String("db_path", perm.MenuPath),
		zap.String("final_icon", icon),
		zap.String("final_path", path))

	node := map[string]interface{}{
		"id":        perm.Code,
		"name":      perm.Name,
		"icon":      icon,
		"path":      path,
		"sort":      perm.SortOrder,
		"type":      perm.Type,
		"component": perm.MenuComponent,
		"visible":   perm.MenuVisible,
		"children":  []interface{}{},
	}

	// 查找子节点
	var children []interface{}
	for _, p := range permMap {
		if p.ParentCode == perm.Code && !processed[p.Code] {
			childNode := s.buildMenuTreeNode(p, permMap, processed)
			children = append(children, childNode)
		}
	}

	node["children"] = children
	return node
}

// generateMenuPath 根据权限生成菜单路径
func (s *permissionService) generateMenuPath(perm models.Permission) string {
	pathMap := map[string]string{
		"system_menu":     "/system",
		"tenant_menu":     "/system/tenants",
		"permission_menu": "/system/permissions",
		"user_menu":       "/users",
		"role_menu":       "/roles",
	}

	if path, exists := pathMap[perm.Code]; exists {
		return path
	}

	// 默认路径生成策略
	switch perm.Module {
	case "user":
		return "/users"
	case "role":
		return "/roles"
	case "system":
		return "/system"
	case "tenant":
		return "/system/tenants"
	default:
		return "/" + perm.Module
	}
}

// generateMenuIcon 根据权限生成菜单图标
func (s *permissionService) generateMenuIcon(perm models.Permission) string {
	iconMap := map[string]string{
		"system_menu":     "settings",
		"tenant_menu":     "apartment",
		"permission_menu": "security",
		"user_menu":       "person",
		"role_menu":       "group",
	}

	if icon, exists := iconMap[perm.Code]; exists {
		return icon
	}

	// 默认图标生成策略
	switch perm.Module {
	case "user":
		return "person"
	case "role":
		return "group"
	case "system":
		return "settings"
	case "tenant":
		return "apartment"
	default:
		return "menu"
	}
}

// GetMenuPermissionsByScope 根据范围获取菜单权限
func (s *permissionService) GetMenuPermissionsByScope(ctx context.Context, scope string) ([]models.Permission, error) {
	s.logger.DebugWithTrace(ctx, "Getting menu permissions by scope",
		zap.String("scope", scope))

	filter := map[string]interface{}{
		"type":  "menu",
		"scope": scope,
	}

	permissions, _, err := s.permissionRepo.List(ctx, filter, 1, 1000) // 获取所有菜单权限
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get menu permissions by scope",
			zap.String("scope", scope),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get menu permissions: %w", err)
	}

	s.logger.DebugWithTrace(ctx, "Retrieved menu permissions by scope",
		zap.String("scope", scope),
		zap.Int("count", len(permissions)))

	return permissions, nil
}
