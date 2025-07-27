// Package services contains business logic and service layer implementations.
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

// RoleService 角色服务接口
type RoleService interface {
	// CreateRole 创建角色
	CreateRole(ctx context.Context, role *models.Role) (*models.Role, error)
	// GetRoleByID 根据ID获取角色
	GetRoleByID(ctx context.Context, id uint64) (*models.Role, error)
	// GetRoleByCode 根据代码获取角色
	GetRoleByCode(ctx context.Context, code, tenantID string) (*models.Role, error)
	// ListRoles 获取角色列表
	ListRoles(ctx context.Context, tenantID string, page, limit int) ([]models.Role, int64, error)
	// UpdateRole 更新角色
	UpdateRole(ctx context.Context, role *models.Role) (*models.Role, error)
	// DeleteRole 删除角色
	DeleteRole(ctx context.Context, id uint64) error
	// AssignPermissions 分配权限给角色
	AssignPermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error
	// RemovePermission 从角色移除权限
	RemovePermission(ctx context.Context, roleID, permissionID uint64) error
	// GetRolePermissions 获取角色权限
	GetRolePermissions(ctx context.Context, roleID uint64) ([]models.Permission, error)
}

// roleService 角色服务实现
type roleService struct {
	roleRepo       repositories.RoleRepository
	permissionRepo repositories.PermissionRepository
	logger         *logger.Logger
}

// NewRoleService 创建角色服务
func NewRoleService(
	roleRepo repositories.RoleRepository,
	permissionRepo repositories.PermissionRepository,
	logger *logger.Logger,
) RoleService {
	return &roleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		logger:         logger,
	}
}

// CreateRole 创建角色
func (s *roleService) CreateRole(ctx context.Context, role *models.Role) (*models.Role, error) {
	s.logger.DebugWithTrace(ctx, "Creating role",
		zap.String("code", role.Code),
		zap.String("name", role.Name),
		zap.Uint64("tenant_id", role.TenantID))

	// 检查角色代码是否已存在
	existingRole, err := s.roleRepo.GetByCode(ctx, role.TenantID, role.Code)
	if err == nil && existingRole != nil {
		return nil, fmt.Errorf("role with code %s already exists in this tenant", role.Code)
	}

	// 创建角色
	if err := s.roleRepo.Create(ctx, role); err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to create role",
			zap.Error(err),
			zap.String("code", role.Code))
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	s.logger.InfoWithTrace(ctx, "Role created successfully",
		zap.String("code", role.Code),
		zap.Uint64("role_id", role.ID))

	return role, nil
}

// GetRoleByID 根据ID获取角色
func (s *roleService) GetRoleByID(ctx context.Context, id uint64) (*models.Role, error) {
	s.logger.DebugWithTrace(ctx, "Getting role by ID",
		zap.Uint64("role_id", id))

	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get role by ID",
			zap.Error(err),
			zap.Uint64("role_id", id))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return role, nil
}

// GetRoleByCode 根据代码获取角色
func (s *roleService) GetRoleByCode(ctx context.Context, code, tenantID string) (*models.Role, error) {
	s.logger.DebugWithTrace(ctx, "Getting role by code",
		zap.String("code", code),
		zap.String("tenant_id", tenantID))

	// 转换tenantID为uint64
	tenantIDUint64, err := s.convertTenantID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	role, err := s.roleRepo.GetByCode(ctx, tenantIDUint64, code)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get role by code",
			zap.Error(err),
			zap.String("code", code))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return role, nil
}

// ListRoles 获取角色列表
func (s *roleService) ListRoles(ctx context.Context, tenantID string, page, limit int) ([]models.Role, int64, error) {
	s.logger.DebugWithTrace(ctx, "Listing roles",
		zap.String("tenant_id", tenantID),
		zap.Int("page", page),
		zap.Int("limit", limit))

	// 转换tenantID为uint64
	tenantIDUint64, err := s.convertTenantID(tenantID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid tenant ID: %w", err)
	}

	// 获取租户角色列表
	allRoles, err := s.roleRepo.GetTenantRoles(ctx, tenantIDUint64)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to list roles",
			zap.Error(err),
			zap.String("tenant_id", tenantID))
		return nil, 0, fmt.Errorf("failed to list roles: %w", err)
	}

	// 实现分页逻辑
	total := int64(len(allRoles))
	start := (page - 1) * limit
	end := start + limit

	if start >= len(allRoles) {
		return []models.Role{}, total, nil
	}
	if end > len(allRoles) {
		end = len(allRoles)
	}

	roles := allRoles[start:end]

	s.logger.DebugWithTrace(ctx, "Retrieved roles",
		zap.Int("count", len(roles)),
		zap.Int64("total", total))

	return roles, total, nil
}

// UpdateRole 更新角色
func (s *roleService) UpdateRole(ctx context.Context, role *models.Role) (*models.Role, error) {
	s.logger.DebugWithTrace(ctx, "Updating role",
		zap.Uint64("role_id", role.ID),
		zap.String("name", role.Name))

	// 检查角色是否存在
	existingRole, err := s.roleRepo.GetByID(ctx, role.ID)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// 检查代码是否与其他角色冲突
	if existingRole.Code != role.Code {
		conflictRole, err := s.roleRepo.GetByCode(ctx, role.TenantID, role.Code)
		if err == nil && conflictRole != nil && conflictRole.ID != role.ID {
			return nil, fmt.Errorf("role with code %s already exists in this tenant", role.Code)
		}
	}

	// 更新角色
	if err := s.roleRepo.Update(ctx, role); err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to update role",
			zap.Error(err),
			zap.Uint64("role_id", role.ID))
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	s.logger.InfoWithTrace(ctx, "Role updated successfully",
		zap.Uint64("role_id", role.ID))

	return role, nil
}

// DeleteRole 删除角色
func (s *roleService) DeleteRole(ctx context.Context, id uint64) error {
	s.logger.DebugWithTrace(ctx, "Deleting role",
		zap.Uint64("role_id", id))

	// 检查角色是否存在
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// 检查是否为系统内置角色
	if role.Type == "system" {
		return fmt.Errorf("cannot delete system built-in role")
	}

	// 删除角色
	if err := s.roleRepo.Delete(ctx, id); err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to delete role",
			zap.Error(err),
			zap.Uint64("role_id", id))
		return fmt.Errorf("failed to delete role: %w", err)
	}

	s.logger.InfoWithTrace(ctx, "Role deleted successfully",
		zap.Uint64("role_id", id))

	return nil
}

// AssignPermissions 分配权限给角色
func (s *roleService) AssignPermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error {
	s.logger.DebugWithTrace(ctx, "Assigning permissions to role",
		zap.Uint64("role_id", roleID),
		zap.Ints("permission_ids", func() []int {
			result := make([]int, len(permissionIDs))
			for i, id := range permissionIDs {
				result[i] = int(id)
			}
			return result
		}()))

	// 检查角色是否存在
	_, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// 验证权限是否存在
	for _, permissionID := range permissionIDs {
		_, err := s.permissionRepo.GetByNumericID(ctx, permissionID)
		if err != nil {
			s.logger.WarnWithTrace(ctx, "Permission not found, skipping",
				zap.Uint64("permission_id", permissionID),
				zap.Error(err))
			continue
		}
	}

	// 分配新权限
	successCount := 0
	for _, permissionID := range permissionIDs {
		rolePermission := &models.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
		}

		if err := s.permissionRepo.AssignPermissionToRole(ctx, rolePermission); err != nil {
			s.logger.WarnWithTrace(ctx, "Failed to assign permission to role, may already exist",
				zap.Error(err),
				zap.Uint64("role_id", roleID),
				zap.Uint64("permission_id", permissionID))
			continue
		}
		successCount++
	}

	s.logger.InfoWithTrace(ctx, "Permissions assigned to role",
		zap.Uint64("role_id", roleID),
		zap.Int("total_permissions", len(permissionIDs)),
		zap.Int("successfully_assigned", successCount))

	return nil
}

// RemovePermission 从角色移除权限
func (s *roleService) RemovePermission(ctx context.Context, roleID, permissionID uint64) error {
	s.logger.DebugWithTrace(ctx, "Removing permission from role",
		zap.Uint64("role_id", roleID),
		zap.Uint64("permission_id", permissionID))

	if err := s.permissionRepo.RemovePermissionFromRole(ctx, fmt.Sprintf("%d", roleID), fmt.Sprintf("%d", permissionID)); err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to remove permission from role",
			zap.Error(err),
			zap.Uint64("role_id", roleID),
			zap.Uint64("permission_id", permissionID))
		return fmt.Errorf("failed to remove permission: %w", err)
	}

	s.logger.InfoWithTrace(ctx, "Permission removed from role successfully",
		zap.Uint64("role_id", roleID),
		zap.Uint64("permission_id", permissionID))

	return nil
}

// GetRolePermissions 获取角色权限
func (s *roleService) GetRolePermissions(ctx context.Context, roleID uint64) ([]models.Permission, error) {
	s.logger.DebugWithTrace(ctx, "Getting role permissions",
		zap.Uint64("role_id", roleID))

	permissions, err := s.permissionRepo.GetRolePermissions(ctx, fmt.Sprintf("%d", roleID))
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get role permissions",
			zap.Error(err),
			zap.Uint64("role_id", roleID))
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	s.logger.DebugWithTrace(ctx, "Retrieved role permissions",
		zap.Uint64("role_id", roleID),
		zap.Int("permission_count", len(permissions)))

	return permissions, nil
}

// convertTenantID 转换tenantID为uint64
func (s *roleService) convertTenantID(tenantID string) (uint64, error) {
	// 如果为空字符串，返回0（系统租户）
	if tenantID == "" || tenantID == "0" {
		return 0, nil
	}

	// 检查是否为数字格式
	if numericID, err := strconv.ParseUint(tenantID, 10, 64); err == nil {
		return numericID, nil
	}

	// 检查是否为UUID格式
	if len(tenantID) == 36 && tenantID[8] == '-' && tenantID[13] == '-' {
		// TODO: 这里应该查询数据库将UUID转换为数字ID
		// 暂时简单映射
		switch tenantID {
		case "90e3fe3e-58a4-11f0-af3a-eeae1ed9f0ce":
			return 1, nil
		default:
			return 0, fmt.Errorf("unknown tenant UUID: %s", tenantID)
		}
	}

	return 0, fmt.Errorf("invalid tenant ID format: %s", tenantID)
}
