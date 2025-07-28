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

//go:generate mockgen -source=role_repository.go -destination=mocks/role_repository_mock.go

// RoleRepository 角色仓储接口
type RoleRepository interface {
	// GetByID 根据ID获取角色（内部调用）
	GetByID(ctx context.Context, id uint64) (*models.Role, error)
	// GetByUUID 根据UUID获取角色（对外接口调用）
	GetByUUID(ctx context.Context, uuid string) (*models.Role, error)
	// GetByCode 根据代码获取角色
	GetByCode(ctx context.Context, tenantID uint64, code string) (*models.Role, error)
	// GetUserRoles 获取用户的角色列表
	GetUserRoles(ctx context.Context, userID, tenantID uint64) ([]models.Role, error)
	// GetUserRolesByUUID 获取用户的角色列表（通过UUID）
	GetUserRolesByUUID(ctx context.Context, userUUID, tenantUUID string) ([]models.Role, error)
	// GetTenantRoles 获取租户的角色列表
	GetTenantRoles(ctx context.Context, tenantID uint64) ([]models.Role, error)
	// Create 创建角色
	Create(ctx context.Context, role *models.Role) error
	// Update 更新角色
	Update(ctx context.Context, role *models.Role) error
	// Delete 删除角色
	Delete(ctx context.Context, id uint64) error
	// DeleteByUUID 删除角色（通过UUID）
	DeleteByUUID(ctx context.Context, uuid string) error
	// AssignRoleToUser 将角色分配给用户
	AssignRoleToUser(ctx context.Context, userRole *models.UserRole) error
	// RemoveRoleFromUser 从用户移除角色
	RemoveRoleFromUser(ctx context.Context, userID, roleID uint64) error
	// RemoveRoleFromUserByUUID 从用户移除角色（通过UUID）
	RemoveRoleFromUserByUUID(ctx context.Context, userUUID, roleUUID string) error
	// List 获取角色列表（支持过滤和分页）
	List(ctx context.Context, filter map[string]interface{}, page, limit int) ([]models.Role, int64, error)
	// GetRolesByScope 根据作用域获取角色列表
	GetRolesByScope(ctx context.Context, tenantID uint64, roleType string) ([]models.Role, error)
	// GetSystemRoles 获取系统角色列表
	GetSystemRoles(ctx context.Context) ([]models.Role, error)
}

// roleRepository 角色仓储实现
type roleRepository struct {
	*transaction.BaseRepository
	logger *logger.Logger
}

// NewRoleRepository 创建角色仓储
func NewRoleRepository(db *gorm.DB, txManager transaction.TransactionManager, logger *logger.Logger) RoleRepository {
	return &roleRepository{
		BaseRepository: transaction.NewBaseRepository(db, txManager, logger.Logger),
		logger:         logger,
	}
}

// GetByID 根据ID获取角色（内部调用）
func (r *roleRepository) GetByID(ctx context.Context, id uint64) (*models.Role, error) {
	var role models.Role

	r.logger.DebugWithTrace(ctx, "Getting role by ID",
		zap.Uint64("role_id", id))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).First(&role, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnWithTrace(ctx, "Role not found by ID",
				zap.Uint64("role_id", id))
			return nil, fmt.Errorf("role not found")
		}
		r.logger.ErrorWithTrace(ctx, "Failed to get role by ID",
			zap.Error(err),
			zap.Uint64("role_id", id))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

// GetByUUID 根据UUID获取角色（对外接口调用）
func (r *roleRepository) GetByUUID(ctx context.Context, uuid string) (*models.Role, error) {
	var role models.Role

	r.logger.DebugWithTrace(ctx, "Getting role by UUID",
		zap.String("role_uuid", uuid))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Where("uuid = ?", uuid).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnWithTrace(ctx, "Role not found by UUID",
				zap.String("role_uuid", uuid))
			return nil, fmt.Errorf("role not found")
		}
		r.logger.ErrorWithTrace(ctx, "Failed to get role by UUID",
			zap.Error(err),
			zap.String("role_uuid", uuid))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

// GetByCode 根据代码获取角色
func (r *roleRepository) GetByCode(ctx context.Context, tenantID uint64, code string) (*models.Role, error) {
	var role models.Role

	r.logger.DebugWithTrace(ctx, "Getting role by code",
		zap.Uint64("tenant_id", tenantID),
		zap.String("code", code))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Where("tenant_id = ? AND code = ?", tenantID, code).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WarnWithTrace(ctx, "Role not found by code",
				zap.Uint64("tenant_id", tenantID),
				zap.String("code", code))
			return nil, fmt.Errorf("role not found")
		}
		r.logger.ErrorWithTrace(ctx, "Failed to get role by code",
			zap.Error(err),
			zap.Uint64("tenant_id", tenantID),
			zap.String("code", code))
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

// GetUserRoles 获取用户的角色列表
func (r *roleRepository) GetUserRoles(ctx context.Context, userID, tenantID uint64) ([]models.Role, error) {
	var roles []models.Role

	r.logger.DebugWithTrace(ctx, "Getting user roles by ID",
		zap.Uint64("user_id", userID),
		zap.Uint64("tenant_id", tenantID))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).
		Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND user_roles.tenant_id = ? AND user_roles.is_active = ?", userID, tenantID, true).
		Where("roles.is_active = ?", true).
		Find(&roles).Error

	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to get user roles by ID",
			zap.Error(err),
			zap.Uint64("user_id", userID),
			zap.Uint64("tenant_id", tenantID))
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved user roles by ID",
		zap.Uint64("user_id", userID),
		zap.Int("role_count", len(roles)))

	return roles, nil
}

// GetUserRolesByUUID 获取用户的角色列表（通过UUID）
func (r *roleRepository) GetUserRolesByUUID(ctx context.Context, userUUID, tenantUUID string) ([]models.Role, error) {
	var roles []models.Role

	r.logger.DebugWithTrace(ctx, "Getting user roles by UUID",
		zap.String("user_uuid", userUUID),
		zap.String("tenant_uuid", tenantUUID))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).
		Table("roles").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Joins("JOIN users ON user_roles.user_id = users.id").
		Joins("JOIN tenants ON user_roles.tenant_id = tenants.id").
		Where("users.uuid = ? AND tenants.uuid = ? AND user_roles.is_active = ?", userUUID, tenantUUID, true).
		Where("roles.is_active = ?", true).
		Find(&roles).Error

	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to get user roles by UUID",
			zap.Error(err),
			zap.String("user_uuid", userUUID),
			zap.String("tenant_uuid", tenantUUID))
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved user roles by UUID",
		zap.String("user_uuid", userUUID),
		zap.Int("role_count", len(roles)))

	return roles, nil
}

// GetTenantRoles 获取租户的角色列表
func (r *roleRepository) GetTenantRoles(ctx context.Context, tenantID uint64) ([]models.Role, error) {
	var roles []models.Role

	r.logger.DebugWithTrace(ctx, "Getting tenant roles",
		zap.Uint64("tenant_id", tenantID))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Where("tenant_id = ? AND is_active = ?", tenantID, true).Find(&roles).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to get tenant roles",
			zap.Error(err),
			zap.Uint64("tenant_id", tenantID))
		return nil, fmt.Errorf("failed to get tenant roles: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved tenant roles",
		zap.Uint64("tenant_id", tenantID),
		zap.Int("role_count", len(roles)))

	return roles, nil
}

// Create 创建角色
func (r *roleRepository) Create(ctx context.Context, role *models.Role) error {
	r.logger.DebugWithTrace(ctx, "Creating role",
		zap.String("code", role.Code),
		zap.String("name", role.Name),
		zap.Uint64("tenant_id", role.TenantID))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Create(role).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to create role",
			zap.Error(err),
			zap.String("code", role.Code))
		return fmt.Errorf("failed to create role: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Role created",
		zap.Uint64("role_id", role.ID),
		zap.String("role_uuid", role.UUID),
		zap.String("code", role.Code))

	return nil
}

// Update 更新角色
func (r *roleRepository) Update(ctx context.Context, role *models.Role) error {
	r.logger.DebugWithTrace(ctx, "Updating role",
		zap.Uint64("role_id", role.ID),
		zap.String("role_uuid", role.UUID),
		zap.String("code", role.Code))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Save(role).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to update role",
			zap.Error(err),
			zap.Uint64("role_id", role.ID),
			zap.String("role_uuid", role.UUID))
		return fmt.Errorf("failed to update role: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Role updated",
		zap.Uint64("role_id", role.ID),
		zap.String("role_uuid", role.UUID))

	return nil
}

// Delete 删除角色
func (r *roleRepository) Delete(ctx context.Context, id uint64) error {
	r.logger.DebugWithTrace(ctx, "Deleting role by ID",
		zap.Uint64("role_id", id))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Delete(&models.Role{}, id).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to delete role by ID",
			zap.Error(err),
			zap.Uint64("role_id", id))
		return fmt.Errorf("failed to delete role: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Role deleted by ID",
		zap.Uint64("role_id", id))

	return nil
}

// DeleteByUUID 删除角色（通过UUID）
func (r *roleRepository) DeleteByUUID(ctx context.Context, uuid string) error {
	r.logger.DebugWithTrace(ctx, "Deleting role by UUID",
		zap.String("role_uuid", uuid))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&models.Role{}).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to delete role by UUID",
			zap.Error(err),
			zap.String("role_uuid", uuid))
		return fmt.Errorf("failed to delete role: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Role deleted by UUID",
		zap.String("role_uuid", uuid))

	return nil
}

// AssignRoleToUser 将角色分配给用户
func (r *roleRepository) AssignRoleToUser(ctx context.Context, userRole *models.UserRole) error {
	r.logger.DebugWithTrace(ctx, "Assigning role to user",
		zap.Uint64("user_id", userRole.UserID),
		zap.Uint64("role_id", userRole.RoleID),
		zap.Uint64("tenant_id", userRole.TenantID))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Create(userRole).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to assign role to user",
			zap.Error(err),
			zap.Uint64("user_id", userRole.UserID),
			zap.Uint64("role_id", userRole.RoleID))
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Role assigned to user",
		zap.Uint64("user_role_id", userRole.ID),
		zap.Uint64("user_id", userRole.UserID),
		zap.Uint64("role_id", userRole.RoleID))

	return nil
}

// RemoveRoleFromUser 从用户移除角色
func (r *roleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uint64) error {
	r.logger.DebugWithTrace(ctx, "Removing role from user by ID",
		zap.Uint64("user_id", userID),
		zap.Uint64("role_id", roleID))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&models.UserRole{}).Error

	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to remove role from user by ID",
			zap.Error(err),
			zap.Uint64("user_id", userID),
			zap.Uint64("role_id", roleID))
		return fmt.Errorf("failed to remove role from user: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Role removed from user by ID",
		zap.Uint64("user_id", userID),
		zap.Uint64("role_id", roleID))

	return nil
}

// RemoveRoleFromUserByUUID 从用户移除角色（通过UUID）
func (r *roleRepository) RemoveRoleFromUserByUUID(ctx context.Context, userUUID, roleUUID string) error {
	r.logger.DebugWithTrace(ctx, "Removing role from user by UUID",
		zap.String("user_uuid", userUUID),
		zap.String("role_uuid", roleUUID))

	db := r.GetDB(ctx)

	// 先获取用户和角色的内部ID
	var userID, roleID uint64
	if err := db.WithContext(ctx).Model(&models.User{}).Select("id").Where("uuid = ?", userUUID).Scan(&userID).Error; err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if err := db.WithContext(ctx).Model(&models.Role{}).Select("id").Where("uuid = ?", roleUUID).Scan(&roleID).Error; err != nil {
		return fmt.Errorf("failed to find role: %w", err)
	}

	err := db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&models.UserRole{}).Error

	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to remove role from user by UUID",
			zap.Error(err),
			zap.String("user_uuid", userUUID),
			zap.String("role_uuid", roleUUID))
		return fmt.Errorf("failed to remove role from user: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Role removed from user by UUID",
		zap.String("user_uuid", userUUID),
		zap.String("role_uuid", roleUUID))

	return nil
}

// List 获取角色列表（支持过滤和分页）
func (r *roleRepository) List(ctx context.Context, filter map[string]interface{}, page, limit int) ([]models.Role, int64, error) {
	var roles []models.Role
	var total int64

	r.logger.DebugWithTrace(ctx, "Getting roles list",
		zap.Int("page", page),
		zap.Int("limit", limit))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).
		Model(&models.Role{}).
		Where(filter).
		Count(&total).
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&roles).Error

	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to get roles list",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("limit", limit))
		return nil, 0, fmt.Errorf("failed to get roles list: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved roles list",
		zap.Int("role_count", len(roles)),
		zap.Int64("total", total))

	return roles, total, nil
}

// GetRolesByScope 根据作用域获取角色列表
func (r *roleRepository) GetRolesByScope(ctx context.Context, tenantID uint64, roleType string) ([]models.Role, error) {
	var roles []models.Role

	r.logger.DebugWithTrace(ctx, "Getting roles by scope",
		zap.Uint64("tenant_id", tenantID),
		zap.String("role_type", roleType))

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).Where("tenant_id = ? AND role_type = ?", tenantID, roleType).Find(&roles).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to get roles by scope",
			zap.Error(err),
			zap.Uint64("tenant_id", tenantID),
			zap.String("role_type", roleType))
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved roles by scope",
		zap.Uint64("tenant_id", tenantID),
		zap.String("role_type", roleType),
		zap.Int("role_count", len(roles)))

	return roles, nil
}

// GetSystemRoles 获取系统角色列表
func (r *roleRepository) GetSystemRoles(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role

	r.logger.DebugWithTrace(ctx, "Getting system roles")

	db := r.GetDB(ctx)
	err := db.WithContext(ctx).
		Where("tenant_id = 0 AND type = ? AND is_active = ?", "system", true).
		Order("id ASC").
		Find(&roles).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to get system roles",
			zap.Error(err))
		return nil, fmt.Errorf("failed to get system roles: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved system roles",
		zap.Int("role_count", len(roles)))

	return roles, nil
}
