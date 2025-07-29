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

//go:generate mockgen -source=field_permission_repository.go -destination=mocks/field_permission_repository_mock.go

// FieldPermissionRepository 字段权限仓储接口
type FieldPermissionRepository interface {
	// GetFieldPermissions 获取表的字段权限配置
	GetFieldPermissions(ctx context.Context, tableName string) ([]models.FieldPermission, error)
	// GetRoleFieldPermissions 获取角色的字段权限
	GetRoleFieldPermissions(ctx context.Context, roleID uint64, tableName string) ([]models.RoleFieldPermission, error)
	// UpdateRoleFieldPermissions 更新角色的字段权限
	UpdateRoleFieldPermissions(ctx context.Context, roleID uint64, tableName string, permissions []models.RoleFieldPermission) error
	// GetUserFieldPermissions 获取用户的字段权限（通过角色）
	GetUserFieldPermissions(ctx context.Context, userID, tenantID uint64, tableName string) (map[string]string, error)
	// CreateFieldPermission 创建字段权限配置
	CreateFieldPermission(ctx context.Context, permission *models.FieldPermission) error
	// CreateFieldPermissions 批量创建字段权限配置
	CreateFieldPermissions(ctx context.Context, permissions []models.FieldPermission) error
	// UpdateFieldPermission 更新字段权限配置
	UpdateFieldPermission(ctx context.Context, permission *models.FieldPermission) error
	// DeleteFieldPermission 删除字段权限配置
	DeleteFieldPermission(ctx context.Context, id uint64) error
}

// fieldPermissionRepository 字段权限仓储实现
type fieldPermissionRepository struct {
	*transaction.BaseRepository
	logger *logger.Logger
}

// NewFieldPermissionRepository 创建字段权限仓储
func NewFieldPermissionRepository(db *gorm.DB, txManager transaction.TransactionManager, logger *logger.Logger) FieldPermissionRepository {
	return &fieldPermissionRepository{
		BaseRepository: transaction.NewBaseRepository(db, txManager, logger.Logger),
		logger:         logger,
	}
}

// GetFieldPermissions 获取表的字段权限配置
func (r *fieldPermissionRepository) GetFieldPermissions(ctx context.Context, tableName string) ([]models.FieldPermission, error) {
	r.LogTransactionState(ctx, "Get Field Permissions")

	var permissions []models.FieldPermission
	db := r.GetDB(ctx)

	err := db.Where("entity_table = ? AND is_active = ?", tableName, true).
		Order("sort_order ASC, field_name ASC").
		Find(&permissions).Error

	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to get field permissions",
			zap.String("table_name", tableName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get field permissions: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved field permissions",
		zap.String("table_name", tableName),
		zap.Int("count", len(permissions)))

	return permissions, nil
}

// GetRoleFieldPermissions 获取角色的字段权限  
func (r *fieldPermissionRepository) GetRoleFieldPermissions(ctx context.Context, roleID uint64, tableName string) ([]models.RoleFieldPermission, error) {
	r.LogTransactionState(ctx, "Get Role Field Permissions")

	var permissions []models.RoleFieldPermission
	db := r.GetDB(ctx)

	err := db.Where("role_id = ? AND entity_table = ?", roleID, tableName).
		Find(&permissions).Error

	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to get role field permissions",
			zap.Uint64("role_id", roleID),
			zap.String("table_name", tableName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get role field permissions: %w", err)
	}

	r.logger.DebugWithTrace(ctx, "Retrieved role field permissions",
		zap.Uint64("role_id", roleID),
		zap.String("table_name", tableName),
		zap.Int("count", len(permissions)))

	return permissions, nil
}

// UpdateRoleFieldPermissions 更新角色的字段权限
func (r *fieldPermissionRepository) UpdateRoleFieldPermissions(ctx context.Context, roleID uint64, tableName string, permissions []models.RoleFieldPermission) error {
	r.LogTransactionState(ctx, "Update Role Field Permissions")

	db := r.GetDB(ctx)

	// 在事务中执行
	err := db.Transaction(func(tx *gorm.DB) error {
		// 删除现有权限
		err := tx.Where("role_id = ? AND entity_table = ?", roleID, tableName).
			Delete(&models.RoleFieldPermission{}).Error
		if err != nil {
			return fmt.Errorf("failed to delete existing permissions: %w", err)
		}

		// 如果有新权限，则插入
		if len(permissions) > 0 {
			err = tx.Create(&permissions).Error
			if err != nil {
				return fmt.Errorf("failed to create new permissions: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to update role field permissions",
			zap.Uint64("role_id", roleID),
			zap.String("table_name", tableName),
			zap.Error(err))
		return fmt.Errorf("failed to update role field permissions: %w", err)
	}

	r.logger.InfoWithTrace(ctx, "Role field permissions updated successfully",
		zap.Uint64("role_id", roleID),
		zap.String("table_name", tableName),
		zap.Int("permission_count", len(permissions)))

	return nil
}

// GetUserFieldPermissions 获取用户的字段权限（通过角色）
func (r *fieldPermissionRepository) GetUserFieldPermissions(ctx context.Context, userID, tenantID uint64, tableName string) (map[string]string, error) {
	r.LogTransactionState(ctx, "Get User Field Permissions")

	db := r.GetDB(ctx)
	permissions := make(map[string]string)

	// 查询用户的角色字段权限
	query := `
		SELECT rfp.field_name, rfp.permission_type
		FROM role_field_permissions rfp
		INNER JOIN user_roles ur ON rfp.role_id = ur.role_id
		WHERE ur.user_id = ? AND ur.tenant_id = ? AND rfp.entity_table = ? AND ur.is_active = true
	`

	var results []struct {
		FieldName      string `db:"field_name"`
		PermissionType string `db:"permission_type"`
	}

	err := db.Raw(query, userID, tenantID, tableName).Scan(&results).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to get user field permissions",
			zap.Uint64("user_id", userID),
			zap.Uint64("tenant_id", tenantID),
			zap.String("table_name", tableName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user field permissions: %w", err)
	}

	// 构建权限映射
	for _, result := range results {
		// 如果已存在权限，取更高权限（default > readonly > hidden）
		if existing, exists := permissions[result.FieldName]; exists {
			if getPermissionLevel(result.PermissionType) > getPermissionLevel(existing) {
				permissions[result.FieldName] = result.PermissionType
			}
		} else {
			permissions[result.FieldName] = result.PermissionType
		}
	}

	// 如果用户没有特定字段权限，则使用默认权限
	fieldPermissions, err := r.GetFieldPermissions(ctx, tableName)
	if err != nil {
		return permissions, nil // 如果获取默认权限失败，返回已有权限
	}

	for _, field := range fieldPermissions {
		if _, exists := permissions[field.FieldName]; !exists {
			permissions[field.FieldName] = field.DefaultValue
		}
	}

	r.logger.DebugWithTrace(ctx, "Retrieved user field permissions",
		zap.Uint64("user_id", userID),
		zap.String("table_name", tableName),
		zap.Int("permission_count", len(permissions)))

	return permissions, nil
}

// CreateFieldPermission 创建字段权限配置
func (r *fieldPermissionRepository) CreateFieldPermission(ctx context.Context, permission *models.FieldPermission) error {
	r.LogTransactionState(ctx, "Create Field Permission")

	db := r.GetDB(ctx)

	err := db.Create(permission).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to create field permission",
			zap.String("table_name", permission.EntityTable),
			zap.String("field_name", permission.FieldName),
			zap.Error(err))
		return fmt.Errorf("failed to create field permission: %w", err)
	}

	r.logger.InfoWithTrace(ctx, "Field permission created successfully",
		zap.String("table_name", permission.EntityTable),
		zap.String("field_name", permission.FieldName))

	return nil
}

// CreateFieldPermissions 批量创建字段权限配置
func (r *fieldPermissionRepository) CreateFieldPermissions(ctx context.Context, permissions []models.FieldPermission) error {
	r.LogTransactionState(ctx, "Create Field Permissions")

	if len(permissions) == 0 {
		return nil
	}

	db := r.GetDB(ctx)

	err := db.Create(&permissions).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to create field permissions",
			zap.Int("permission_count", len(permissions)),
			zap.Error(err))
		return fmt.Errorf("failed to create field permissions: %w", err)
	}

	r.logger.InfoWithTrace(ctx, "Field permissions created successfully",
		zap.Int("permission_count", len(permissions)))

	return nil
}

// UpdateFieldPermission 更新字段权限配置
func (r *fieldPermissionRepository) UpdateFieldPermission(ctx context.Context, permission *models.FieldPermission) error {
	r.LogTransactionState(ctx, "Update Field Permission")

	db := r.GetDB(ctx)

	err := db.Save(permission).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to update field permission",
			zap.Uint64("id", permission.ID),
			zap.String("table_name", permission.EntityTable),
			zap.String("field_name", permission.FieldName),
			zap.Error(err))
		return fmt.Errorf("failed to update field permission: %w", err)
	}

	r.logger.InfoWithTrace(ctx, "Field permission updated successfully",
		zap.Uint64("id", permission.ID),
		zap.String("table_name", permission.EntityTable),
		zap.String("field_name", permission.FieldName))

	return nil
}

// DeleteFieldPermission 删除字段权限配置
func (r *fieldPermissionRepository) DeleteFieldPermission(ctx context.Context, id uint64) error {
	r.LogTransactionState(ctx, "Delete Field Permission")

	db := r.GetDB(ctx)

	err := db.Delete(&models.FieldPermission{}, id).Error
	if err != nil {
		r.logger.ErrorWithTrace(ctx, "Failed to delete field permission",
			zap.Uint64("id", id),
			zap.Error(err))
		return fmt.Errorf("failed to delete field permission: %w", err)
	}

	r.logger.InfoWithTrace(ctx, "Field permission deleted successfully",
		zap.Uint64("id", id))

	return nil
}

// getPermissionLevel 获取权限级别（用于比较权限优先级）
// default(3) > readonly(2) > hidden(1)
func getPermissionLevel(permissionType string) int {
	switch permissionType {
	case models.FieldPermissionDefault:
		return 3
	case models.FieldPermissionReadonly:
		return 2
	case models.FieldPermissionHidden:
		return 1
	default:
		return 0
	}
}