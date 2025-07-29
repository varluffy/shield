// Package services contains business logic and service layer implementations.
package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/varluffy/shield/internal/dto"
	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/internal/repositories"
	"github.com/varluffy/shield/pkg/logger"
	"go.uber.org/zap"
)

// FieldPermissionService 字段权限服务接口
type FieldPermissionService interface {
	// GetTableFields 获取表的字段配置
	GetTableFields(ctx context.Context, tableName string) ([]models.FieldPermission, error)
	// GetRoleFieldPermissions 获取角色的字段权限
	GetRoleFieldPermissions(ctx context.Context, roleID uint64, tableName string) ([]models.RoleFieldPermission, error)
	// UpdateRoleFieldPermissions 更新角色的字段权限
	UpdateRoleFieldPermissions(ctx context.Context, roleID uint64, tableName string, permissions []models.RoleFieldPermission) error
	// GetUserFieldPermissions 获取用户的字段权限
	GetUserFieldPermissions(ctx context.Context, userID, tenantID, tableName string) (map[string]string, error)
	// InitializeFieldPermissions 初始化表的字段权限配置
	InitializeFieldPermissions(ctx context.Context, tableName string, fields []dto.FieldConfig) error
}

// fieldPermissionService 字段权限服务实现
type fieldPermissionService struct {
	fieldPermissionRepo repositories.FieldPermissionRepository
	userRepo            repositories.UserRepository
	logger              *logger.Logger
}

// NewFieldPermissionService 创建字段权限服务
func NewFieldPermissionService(
	fieldPermissionRepo repositories.FieldPermissionRepository,
	userRepo repositories.UserRepository,
	logger *logger.Logger,
) FieldPermissionService {
	return &fieldPermissionService{
		fieldPermissionRepo: fieldPermissionRepo,
		userRepo:            userRepo,  
		logger:              logger,
	}
}

// GetTableFields 获取表的字段配置
func (s *fieldPermissionService) GetTableFields(ctx context.Context, tableName string) ([]models.FieldPermission, error) {
	s.logger.DebugWithTrace(ctx, "Getting table fields",
		zap.String("table_name", tableName))

	// 从数据库获取字段配置
	fields, err := s.fieldPermissionRepo.GetFieldPermissions(ctx, tableName)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get table fields from repository",
			zap.String("table_name", tableName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get table fields: %w", err)
	}

	// 如果数据库中没有配置，返回默认配置
	if len(fields) == 0 {
		fields = s.getDefaultFieldPermissions(tableName)
		s.logger.WarnWithTrace(ctx, "No field permissions found in database, using default",
			zap.String("table_name", tableName),
			zap.Int("default_count", len(fields)))
	}

	s.logger.DebugWithTrace(ctx, "Retrieved table fields",
		zap.String("table_name", tableName),
		zap.Int("field_count", len(fields)))

	return fields, nil
}

// getDefaultFieldPermissions 获取默认字段权限配置（当数据库中没有配置时使用）
func (s *fieldPermissionService) getDefaultFieldPermissions(tableName string) []models.FieldPermission {
	switch tableName {
	case "users":
		return []models.FieldPermission{
			{
				EntityTable:  "users",
				FieldName:    "name",
				FieldLabel:   "姓名",
				FieldType:    "string",
				DefaultValue: models.FieldPermissionDefault,
				Description:  "用户姓名",
				SortOrder:    1,
				IsActive:     true,
			},
			{
				EntityTable:  "users",
				FieldName:    "email",
				FieldLabel:   "邮箱",
				FieldType:    "email",
				DefaultValue: models.FieldPermissionDefault,
				Description:  "用户邮箱地址",
				SortOrder:    2,
				IsActive:     true,
			},
			{
				EntityTable:  "users",
				FieldName:    "phone",
				FieldLabel:   "手机号",
				FieldType:    "phone",
				DefaultValue: models.FieldPermissionDefault,
				Description:  "手机号码",
				SortOrder:    3,
				IsActive:     true,
			},
			{
				EntityTable:  "users",
				FieldName:    "salary",
				FieldLabel:   "薪资",
				FieldType:    "decimal",
				DefaultValue: models.FieldPermissionHidden,
				Description:  "员工薪资信息",
				SortOrder:    4,
				IsActive:     true,
			},
		}
	case "candidates":
		return []models.FieldPermission{
			{
				EntityTable:  "candidates",
				FieldName:    "name",
				FieldLabel:   "候选人姓名",
				FieldType:    "string",
				DefaultValue: models.FieldPermissionDefault,
				Description:  "候选人姓名",
				SortOrder:    1,
				IsActive:     true,
			},
			{
				EntityTable:  "candidates",
				FieldName:    "resume",
				FieldLabel:   "简历",
				FieldType:    "text",
				DefaultValue: models.FieldPermissionDefault,
				Description:  "候选人简历",
				SortOrder:    2,
				IsActive:     true,
			},
			{
				EntityTable:  "candidates",
				FieldName:    "salary_expectation",
				FieldLabel:   "期望薪资",
				FieldType:    "decimal",
				DefaultValue: models.FieldPermissionReadonly,
				Description:  "候选人期望薪资",
				SortOrder:    3,
				IsActive:     true,
			},
		}
	default:
		return []models.FieldPermission{}
	}
}

// GetRoleFieldPermissions 获取角色的字段权限
func (s *fieldPermissionService) GetRoleFieldPermissions(ctx context.Context, roleID uint64, tableName string) ([]models.RoleFieldPermission, error) {
	s.logger.DebugWithTrace(ctx, "Getting role field permissions",
		zap.Uint64("role_id", roleID),
		zap.String("table_name", tableName))

	// 从数据库获取角色字段权限
	permissions, err := s.fieldPermissionRepo.GetRoleFieldPermissions(ctx, roleID, tableName)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get role field permissions from repository",
			zap.Uint64("role_id", roleID),
			zap.String("table_name", tableName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get role field permissions: %w", err)
	}

	s.logger.DebugWithTrace(ctx, "Retrieved role field permissions",
		zap.Uint64("role_id", roleID),
		zap.String("table_name", tableName),
		zap.Int("permission_count", len(permissions)))

	return permissions, nil
}

// UpdateRoleFieldPermissions 更新角色的字段权限
func (s *fieldPermissionService) UpdateRoleFieldPermissions(ctx context.Context, roleID uint64, tableName string, permissions []models.RoleFieldPermission) error {
	s.logger.DebugWithTrace(ctx, "Updating role field permissions",
		zap.Uint64("role_id", roleID),
		zap.String("table_name", tableName),
		zap.Int("permission_count", len(permissions)))

	// 验证权限类型
	for _, perm := range permissions {
		if !isValidPermissionType(perm.PermissionType) {
			return fmt.Errorf("invalid permission type: %s", perm.PermissionType)
		}
	}

	// 调用Repository更新权限
	err := s.fieldPermissionRepo.UpdateRoleFieldPermissions(ctx, roleID, tableName, permissions)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to update role field permissions in repository",
			zap.Uint64("role_id", roleID),
			zap.String("table_name", tableName),
			zap.Error(err))
		return fmt.Errorf("failed to update role field permissions: %w", err)
	}

	s.logger.InfoWithTrace(ctx, "Role field permissions updated successfully",
		zap.Uint64("role_id", roleID),
		zap.String("table_name", tableName),
		zap.Int("permission_count", len(permissions)))

	return nil
}

// GetUserFieldPermissions 获取用户的字段权限
func (s *fieldPermissionService) GetUserFieldPermissions(ctx context.Context, userID, tenantID, tableName string) (map[string]string, error) {
	s.logger.DebugWithTrace(ctx, "Getting user field permissions",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.String("table_name", tableName))

	// 转换ID为uint64
	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %s", userID)
	}

	tenantIDUint, err := strconv.ParseUint(tenantID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %s", tenantID)
	}

	// 从Repository获取用户字段权限
	permissions, err := s.fieldPermissionRepo.GetUserFieldPermissions(ctx, userIDUint, tenantIDUint, tableName)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to get user field permissions from repository",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.String("table_name", tableName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user field permissions: %w", err)
	}

	s.logger.DebugWithTrace(ctx, "Retrieved user field permissions",
		zap.String("user_id", userID),
		zap.String("table_name", tableName),
		zap.Int("permission_count", len(permissions)))

	return permissions, nil
}

// InitializeFieldPermissions 初始化表的字段权限配置
func (s *fieldPermissionService) InitializeFieldPermissions(ctx context.Context, tableName string, fields []dto.FieldConfig) error {
	s.logger.DebugWithTrace(ctx, "Initializing field permissions",
		zap.String("table_name", tableName),
		zap.Int("field_count", len(fields)))

	// 验证字段配置
	for _, field := range fields {
		if !isValidPermissionType(field.DefaultValue) {
			return fmt.Errorf("invalid default permission type for field %s: %s", field.FieldName, field.DefaultValue)
		}
	}

	// 转换为模型
	var permissions []models.FieldPermission
	for _, field := range fields {
		permissions = append(permissions, models.FieldPermission{
			EntityTable:  tableName,
			FieldName:    field.FieldName,
			FieldLabel:   field.FieldLabel,
			FieldType:    field.FieldType,
			DefaultValue: field.DefaultValue,
			Description:  field.Description,
			SortOrder:    field.SortOrder,
			IsActive:     true,
		})
	}

	// 批量创建字段权限配置
	err := s.fieldPermissionRepo.CreateFieldPermissions(ctx, permissions)
	if err != nil {
		s.logger.ErrorWithTrace(ctx, "Failed to initialize field permissions in repository",
			zap.String("table_name", tableName),
			zap.Error(err))
		return fmt.Errorf("failed to initialize field permissions: %w", err)
	}

	s.logger.InfoWithTrace(ctx, "Field permissions initialized successfully",
		zap.String("table_name", tableName),
		zap.Int("field_count", len(fields)))

	return nil
}

// isValidPermissionType 验证权限类型是否有效
func isValidPermissionType(permType string) bool {
	switch permType {
	case models.FieldPermissionDefault, models.FieldPermissionHidden, models.FieldPermissionReadonly:
		return true
	default:
		return false
	}
}
