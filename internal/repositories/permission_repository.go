// Package repositories contains data access layer implementations.
// It provides repository pattern implementations for database operations.
package repositories

import (
	"context"
	"fmt"

	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/transaction"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//go:generate mockgen -source=permission_repository.go -destination=mocks/permission_repository_mock.go

// PermissionRepository 权限仓储接口
type PermissionRepository interface {
	// GetByID 根据ID获取权限
	GetByID(ctx context.Context, id string) (*models.Permission, error)
	// GetByNumericID 根据数字ID获取权限
	GetByNumericID(ctx context.Context, id uint64) (*models.Permission, error)
	// GetByCode 根据代码获取权限
	GetByCode(ctx context.Context, code string) (*models.Permission, error)
	// GetUserPermissions 获取用户的权限列表
	GetUserPermissions(ctx context.Context, userID, tenantID string) ([]models.Permission, error)
	// GetRolePermissions 获取角色的权限列表
	GetRolePermissions(ctx context.Context, roleID string) ([]models.Permission, error)
	// GetSystemPermissions 获取系统权限列表
	GetSystemPermissions(ctx context.Context) ([]models.Permission, error)
	// List 获取权限列表（支持过滤和分页）
	List(ctx context.Context, filter map[string]interface{}, page, limit int) ([]models.Permission, int64, error)
	// Create 创建权限
	Create(ctx context.Context, permission *models.Permission) error
	// Update 更新权限
	Update(ctx context.Context, permission *models.Permission) (*models.Permission, error)
	// Delete 删除权限
	Delete(ctx context.Context, id string) error
	// AssignPermissionToRole 将权限分配给角色
	AssignPermissionToRole(ctx context.Context, rolePermission *models.RolePermission) error
	// RemovePermissionFromRole 从角色移除权限
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error
	// GetPermissionsByScope 根据作用域获取权限列表
	GetPermissionsByScope(ctx context.Context, scope string) ([]models.Permission, error)
	// GetPermissionsByParent 根据父权限代码获取子权限列表
	GetPermissionsByParent(ctx context.Context, parentCode string) ([]models.Permission, error)
	// GetRootPermissions 获取根权限列表（没有父权限的权限）
	GetRootPermissions(ctx context.Context, scope string) ([]models.Permission, error)
}

// permissionRepository 权限仓储实现
type permissionRepository struct {
	*transaction.BaseRepository
	logger *logger.Logger
}

// NewPermissionRepository 创建权限仓储
func NewPermissionRepository(db *gorm.DB, txManager transaction.TransactionManager, logger *logger.Logger) PermissionRepository {
	return &permissionRepository{
		BaseRepository: transaction.NewBaseRepository(db, txManager, logger.Logger),
		logger:         logger,
	}
}

// GetByID 根据ID获取权限
func (r *permissionRepository) GetByID(ctx context.Context, id string) (*models.Permission, error) {
	var permission models.Permission

	r.logger.DebugWithTrace(ctx, "Getting permission by ID",
		zap.String("permission_id", id))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).First(&permission, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnWithTrace(ctx, "Permission not found",
				zap.String("permission_id", id))
			return nil, fmt.Errorf("permission not found")
		}
		r.logger.ErrorWithTrace(ctx, "Failed to get permission by ID",
			zap.Error(err),
			zap.String("permission_id", id))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return &permission, nil
}

// GetByCode 根据代码获取权限
func (r *permissionRepository) GetByCode(ctx context.Context, code string) (*models.Permission, error) {
	var permission models.Permission

	r.logger.DebugWithTrace(ctx, "Getting permission by code",
		zap.String("code", code))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Where("code = ?", code).First(&permission).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnWithTrace(ctx, "Permission not found by code",
				zap.String("code", code))
			return nil, fmt.Errorf("permission not found")
		}
		r.logger.ErrorWithTrace(ctx, "Failed to get permission by code",
			zap.Error(err),
			zap.String("code", code))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return &permission, nil
}

// GetUserPermissions 获取用户的权限列表
func (r *permissionRepository) GetUserPermissions(ctx context.Context, userID, tenantID string) ([]models.Permission, error) {
	var permissions []models.Permission

	r.logger.DebugWithTrace(ctx, "Getting user permissions",
		zap.String("user_uuid", userID),
		zap.String("tenant_uuid", tenantID))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Joins("JOIN users ON user_roles.user_id = users.id").
		Joins("LEFT JOIN tenants ON user_roles.tenant_id = tenants.id").
		Where("users.uuid = ? AND user_roles.is_active = ? AND permissions.is_active = ?", userID, true, true).
		Where("(tenants.uuid = ? OR user_roles.tenant_id = 0)", tenantID). // 支持系统权限（tenant_id=0）
		Group("permissions.id").
		Find(&permissions).Error

	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to get user permissions",
			zap.Error(err),
			zap.String("user_uuid", userID),
			zap.String("tenant_uuid", tenantID))
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved user permissions",
		zap.String("user_uuid", userID),
		zap.Int("permission_count", len(permissions)))

	return permissions, nil
}

// GetRolePermissions 获取角色的权限列表
func (r *permissionRepository) GetRolePermissions(ctx context.Context, roleID string) ([]models.Permission, error) {
	var permissions []models.Permission

	r.logger.DebugWithTrace(ctx, "Getting role permissions",
		zap.String("role_id", roleID))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error

	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to get role permissions",
			zap.Error(err),
			zap.String("role_id", roleID))
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved role permissions",
		zap.String("role_id", roleID),
		zap.Int("permission_count", len(permissions)))

	return permissions, nil
}

// GetSystemPermissions 获取系统权限列表
func (r *permissionRepository) GetSystemPermissions(ctx context.Context) ([]models.Permission, error) {
	var permissions []models.Permission

	r.logger.DebugWithTrace(ctx, "Getting system permissions")

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).
		Where("scope = ? AND is_active = ?", "system", true).
		Order("sort_order ASC, id ASC").
		Find(&permissions).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to get system permissions",
			zap.Error(err))
		return nil, fmt.Errorf("failed to get system permissions: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved system permissions",
		zap.Int("permission_count", len(permissions)))

	return permissions, nil
}

// Create 创建权限
func (r *permissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	r.logger.DebugWithTrace(ctx, "Creating permission",
		zap.String("code", permission.Code),
		zap.String("name", permission.Name))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Create(permission).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to create permission",
			zap.Error(err),
			zap.String("code", permission.Code))
		return fmt.Errorf("failed to create permission: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Permission created",
		zap.Uint64("permission_id", permission.ID),
		zap.String("code", permission.Code))

	return nil
}

// Update 更新权限
func (r *permissionRepository) Update(ctx context.Context, permission *models.Permission) (*models.Permission, error) {
	r.logger.DebugWithTrace(ctx, "Updating permission",
		zap.Uint64("permission_id", permission.ID),
		zap.String("code", permission.Code))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Save(permission).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to update permission",
			zap.Error(err),
			zap.Uint64("permission_id", permission.ID))
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Permission updated",
		zap.Uint64("permission_id", permission.ID))

	return permission, nil
}

// Delete 删除权限
func (r *permissionRepository) Delete(ctx context.Context, id string) error {
	r.logger.DebugWithTrace(ctx, "Deleting permission",
		zap.String("permission_id", id))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Delete(&models.Permission{}, "id = ?", id).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to delete permission",
			zap.Error(err),
			zap.String("permission_id", id))
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Permission deleted",
		zap.String("permission_id", id))

	return nil
}

// AssignPermissionToRole 将权限分配给角色
func (r *permissionRepository) AssignPermissionToRole(ctx context.Context, rolePermission *models.RolePermission) error {
	r.logger.DebugWithTrace(ctx, "Assigning permission to role",
		zap.Uint64("role_id", rolePermission.RoleID),
		zap.Uint64("permission_id", rolePermission.PermissionID))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Create(rolePermission).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to assign permission to role",
			zap.Error(err),
			zap.Uint64("role_id", rolePermission.RoleID),
			zap.Uint64("permission_id", rolePermission.PermissionID))
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Permission assigned to role",
		zap.Uint64("role_permission_id", rolePermission.ID))

	return nil
}

// RemovePermissionFromRole 从角色移除权限
func (r *permissionRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	r.logger.DebugWithTrace(ctx, "Removing permission from role",
		zap.String("role_id", roleID),
		zap.String("permission_id", permissionID))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Delete(&models.RolePermission{}, "role_id = ? AND permission_id = ?", roleID, permissionID).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to remove permission from role",
			zap.Error(err),
			zap.String("role_id", roleID),
			zap.String("permission_id", permissionID))
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Permission removed from role",
		zap.String("role_id", roleID),
		zap.String("permission_id", permissionID))

	return nil
}

// GetByNumericID 根据数字ID获取权限
func (r *permissionRepository) GetByNumericID(ctx context.Context, id uint64) (*models.Permission, error) {
	var permission models.Permission

	r.logger.DebugWithTrace(ctx, "Getting permission by numeric ID",
		zap.Uint64("permission_id", id))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).First(&permission, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnWithTrace(ctx, "Permission not found",
				zap.Uint64("permission_id", id))
			return nil, fmt.Errorf("permission not found")
		}
		r.logger.ErrorWithTrace(ctx, "Failed to get permission by ID",
			zap.Error(err),
			zap.Uint64("permission_id", id))
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return &permission, nil
}

// List 获取权限列表（支持过滤和分页）
func (r *permissionRepository) List(ctx context.Context, filter map[string]interface{}, page, limit int) ([]models.Permission, int64, error) {
	var permissions []models.Permission
	var total int64

	r.logger.DebugWithTrace(ctx, "Listing permissions with filter",
		zap.Any("filter", filter),
		zap.Int("page", page),
		zap.Int("limit", limit))

	db := r.GetDB(ctx).Model(&models.Permission{})

	// 应用过滤条件
	for key, value := range filter {
		switch key {
		case "scope":
			db = db.Where("scope = ?", value)
		case "module":
			db = db.Where("module = ?", value)
		case "type":
			db = db.Where("type = ?", value)
		case "is_active":
			db = db.Where("is_active = ?", value)
		}
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to count permissions",
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count permissions: %w", err)
	}

	// 分页查询
	offset := (page - 1) * limit
	err := db.Order("sort_order ASC, id ASC").
		Offset(offset).
		Limit(limit).
		Find(&permissions).Error

	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to list permissions",
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list permissions: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved permissions",
		zap.Int("count", len(permissions)),
		zap.Int64("total", total))

	return permissions, total, nil
}

// GetPermissionsByScope 根据作用域获取权限列表
func (r *permissionRepository) GetPermissionsByScope(ctx context.Context, scope string) ([]models.Permission, error) {
	var permissions []models.Permission

	r.logger.DebugWithTrace(ctx, "Getting permissions by scope",
		zap.String("scope", scope))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).
		Where("scope = ? AND is_active = ?", scope, true).
		Order("sort_order ASC, id ASC").
		Find(&permissions).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to get permissions by scope",
			zap.Error(err),
			zap.String("scope", scope))
		return nil, fmt.Errorf("failed to get permissions by scope: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved permissions by scope",
		zap.String("scope", scope),
		zap.Int("permission_count", len(permissions)))

	return permissions, nil
}

// GetPermissionsByParent 根据父权限代码获取子权限列表
func (r *permissionRepository) GetPermissionsByParent(ctx context.Context, parentCode string) ([]models.Permission, error) {
	var permissions []models.Permission

	r.logger.DebugWithTrace(ctx, "Getting permissions by parent",
		zap.String("parent_code", parentCode))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).
		Where("parent_code = ? AND is_active = ?", parentCode, true).
		Order("sort_order ASC, id ASC").
		Find(&permissions).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to get permissions by parent",
			zap.Error(err),
			zap.String("parent_code", parentCode))
		return nil, fmt.Errorf("failed to get permissions by parent: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved permissions by parent",
		zap.String("parent_code", parentCode),
		zap.Int("permission_count", len(permissions)))

	return permissions, nil
}

// GetRootPermissions 获取根权限列表（没有父权限的权限）
func (r *permissionRepository) GetRootPermissions(ctx context.Context, scope string) ([]models.Permission, error) {
	var permissions []models.Permission

	r.logger.DebugWithTrace(ctx, "Getting root permissions",
		zap.String("scope", scope))

	db := r.GetDB(ctx)
	query := db.Where("(parent_code = '' OR parent_code IS NULL) AND is_active = ?", true)
	
	if scope != "" {
		query = query.Where("scope = ?", scope)
	}
	
	err := query.Order("sort_order ASC, id ASC").Find(&permissions).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to get root permissions",
			zap.Error(err),
			zap.String("scope", scope))
		return nil, fmt.Errorf("failed to get root permissions: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved root permissions",
		zap.String("scope", scope),
		zap.Int("permission_count", len(permissions)))

	return permissions, nil
} 