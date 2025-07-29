// Package handlers contains HTTP request handlers for the API endpoints.
package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/varluffy/shield/internal/dto"
	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/internal/services"
	"github.com/varluffy/shield/pkg/errors"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/response"
	"go.uber.org/zap"
)

// FieldPermissionHandler 字段权限处理器
type FieldPermissionHandler struct {
	fieldPermissionService services.FieldPermissionService
	logger                 *logger.Logger
	responseWriter         *response.ResponseWriter
}

// NewFieldPermissionHandler 创建字段权限处理器
func NewFieldPermissionHandler(
	fieldPermissionService services.FieldPermissionService,
	logger *logger.Logger,
) *FieldPermissionHandler {
	return &FieldPermissionHandler{
		fieldPermissionService: fieldPermissionService,
		logger:                 logger,
		responseWriter:         response.NewResponseWriter(logger),
	}
}

// GetFieldMetadata 获取所有可配置的表和字段元数据
// @Summary 获取字段权限元数据
// @Description 获取所有支持字段权限配置的表和字段信息
// @Tags field-permissions
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=dto.FieldMetadataResponse}
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Security BearerAuth
// @Router /field-permissions/metadata [get]
func (h *FieldPermissionHandler) GetFieldMetadata(c *gin.Context) {
	ctx := c.Request.Context()

	h.logger.DebugWithTrace(ctx, "Getting field metadata")

	// 支持字段权限的表列表
	tables := []dto.TableMetadata{
		{
			TableName:   "users",
			TableLabel:  "用户管理",
			Description: "用户基本信息表",
		},
		{
			TableName:   "candidates",
			TableLabel:  "候选人管理",
			Description: "候选人信息表",
		},
		// 可以继续添加其他表
	}

	// 为每个表获取字段信息
	var tableFields []dto.TableFields
	for _, table := range tables {
		fields, err := h.fieldPermissionService.GetTableFields(ctx, table.TableName)
		if err != nil {
			h.logger.ErrorWithTrace(ctx, "Failed to get table fields",
				zap.String("table_name", table.TableName),
				zap.Error(err))
			h.responseWriter.Error(c, errors.ErrInternalError("获取表字段失败"))
			return
		}

		// 转换为DTO
		var fieldDTOs []dto.FieldMetadata
		for _, field := range fields {
			fieldDTOs = append(fieldDTOs, dto.FieldMetadata{
				FieldName:    field.FieldName,
				FieldLabel:   field.FieldLabel,
				FieldType:    field.FieldType,
				Description:  field.Description,
				DefaultValue: field.DefaultValue,
				SortOrder:    field.SortOrder,
				IsActive:     field.IsActive,
			})
		}

		tableFields = append(tableFields, dto.TableFields{
			TableMetadata: table,
			Fields:        fieldDTOs,
		})
	}

	response := dto.FieldMetadataResponse{
		Tables: tableFields,
		PermissionTypes: []dto.PermissionTypeInfo{
			{
				Type:        models.FieldPermissionDefault,
				Label:       "默认",
				Description: "正常显示和编辑",
			},
			{
				Type:        models.FieldPermissionReadonly,
				Label:       "只读",
				Description: "显示但不能编辑",
			},
			{
				Type:        models.FieldPermissionHidden,
				Label:       "隐藏",
				Description: "不显示该字段",
			},
		},
	}

	h.logger.DebugWithTrace(ctx, "Retrieved field metadata",
		zap.Int("table_count", len(tableFields)))

	h.responseWriter.Success(c, response)
}

// GetTableFields 获取表的字段配置
// @Summary 获取表字段配置
// @Description 获取指定表的字段权限配置列表
// @Tags field-permissions
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Security BearerAuth
// @Router /field-permissions/tables/{tableName}/fields [get]
func (h *FieldPermissionHandler) GetTableFields(c *gin.Context) {
	ctx := c.Request.Context()

	tableName := c.Param("tableName")
	if tableName == "" {
		h.logger.WarnWithTrace(ctx, "Missing table name parameter")
		h.responseWriter.Error(c, errors.ErrInvalidRequest())
		return
	}

	fields, err := h.fieldPermissionService.GetTableFields(ctx, tableName)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to get table fields",
			zap.Error(err),
			zap.String("table_name", tableName))
		h.responseWriter.Error(c, errors.ErrInternalError("获取表字段失败"))
		return
	}

	// 转换为DTO格式
	var fieldDTOs []dto.FieldMetadata
	for _, field := range fields {
		fieldDTOs = append(fieldDTOs, dto.FieldMetadata{
			FieldName:    field.FieldName,
			FieldLabel:   field.FieldLabel,
			FieldType:    field.FieldType,
			Description:  field.Description,
			DefaultValue: field.DefaultValue,
			SortOrder:    field.SortOrder,
			IsActive:     field.IsActive,
		})
	}

	response := dto.TableFieldsResponse{
		TableName: tableName,
		Fields:    fieldDTOs,
	}

	h.responseWriter.Success(c, response)
}

// GetRoleFieldPermissions 获取角色的字段权限
// @Summary 获取角色字段权限
// @Description 获取指定角色在指定表的字段权限配置
// @Tags field-permissions
// @Accept json
// @Produce json
// @Param roleId path int true "角色ID"
// @Param tableName path string true "表名"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Security BearerAuth
// @Router /field-permissions/roles/{roleId}/{tableName} [get]
func (h *FieldPermissionHandler) GetRoleFieldPermissions(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取角色ID
	roleIDStr := c.Param("roleId")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid role ID",
			zap.String("role_id", roleIDStr))
		h.responseWriter.BadRequest(c, "Invalid role ID")
		return
	}

	// 获取表名
	tableName := c.Param("tableName")
	if tableName == "" {
		h.logger.WarnWithTrace(ctx, "Missing table name parameter")
		h.responseWriter.BadRequest(c, "Table name is required")
		return
	}

	// 获取角色字段权限
	permissions, err := h.fieldPermissionService.GetRoleFieldPermissions(ctx, roleID, tableName)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to get role field permissions",
			zap.Error(err),
			zap.Uint64("role_id", roleID),
			zap.String("table_name", tableName))
		h.responseWriter.Error(c, err)
		return
	}

	// 同时获取表的字段配置，用于前端展示
	fields, err := h.fieldPermissionService.GetTableFields(ctx, tableName)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to get table fields",
			zap.Error(err),
			zap.String("table_name", tableName))
		h.responseWriter.Error(c, err)
		return
	}

	// 构建权限映射
	permissionMap := make(map[string]string)
	for _, perm := range permissions {
		permissionMap[perm.FieldName] = perm.PermissionType
	}

	// 构建完整的字段权限信息
	var fieldPermissions []dto.FieldPermissionConfig
	for _, field := range fields {
		permission := field.DefaultValue // 使用默认权限
		if p, exists := permissionMap[field.FieldName]; exists {
			permission = p // 使用角色设置的权限
		}

		fieldPermissions = append(fieldPermissions, dto.FieldPermissionConfig{
			FieldName:      field.FieldName,
			FieldLabel:     field.FieldLabel,
			FieldType:      field.FieldType,
			DefaultValue:   field.DefaultValue,
			PermissionType: permission,
			Description:    field.Description,
			SortOrder:      field.SortOrder,
		})
	}

	response := dto.RoleFieldPermissionsResponse{
		RoleID:           roleID,
		TableName:        tableName,
		FieldPermissions: fieldPermissions,
		LastModified:     nil, // TODO: 从数据库获取
	}

	h.responseWriter.Success(c, response)
}

// UpdateRoleFieldPermissions 更新角色的字段权限
// @Summary 更新角色字段权限
// @Description 批量更新角色在指定表的字段权限配置
// @Tags field-permissions
// @Accept json
// @Produce json
// @Param roleId path int true "角色ID"
// @Param tableName path string true "表名"
// @Param permissions body dto.UpdateRoleFieldPermissionsRequest true "字段权限配置"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Security BearerAuth
// @Router /field-permissions/roles/{roleId}/{tableName} [put]
func (h *FieldPermissionHandler) UpdateRoleFieldPermissions(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取角色ID
	roleIDStr := c.Param("roleId")
	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid role ID",
			zap.String("role_id", roleIDStr))
		h.responseWriter.BadRequest(c, "Invalid role ID")
		return
	}

	// 获取表名
	tableName := c.Param("tableName")
	if tableName == "" {
		h.logger.WarnWithTrace(ctx, "Missing table name parameter")
		h.responseWriter.BadRequest(c, "Table name is required")
		return
	}

	// 解析请求体
	var req dto.UpdateRoleFieldPermissionsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid request body for update role field permissions")
		h.responseWriter.ValidationError(c, err)
		return
	}

	// 获取租户ID（用于数据隔离）
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		h.logger.WarnWithTrace(ctx, "Tenant ID not found in context")
		h.responseWriter.Error(c, errors.ErrInternalError("tenant context not found"))
		return
	}

	tenantIDStr, ok := tenantID.(string)
	if !ok {
		h.logger.WarnWithTrace(ctx, "Invalid tenant ID type")
		h.responseWriter.Error(c, errors.ErrInternalError("invalid tenant context"))
		return
	}

	// 转换tenantID为uint64
	tenantIDUint64, err := strconv.ParseUint(tenantIDStr, 10, 64)
	if err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid tenant ID format",
			zap.String("tenant_id", tenantIDStr))
		h.responseWriter.Error(c, errors.ErrInternalError("invalid tenant ID"))
		return
	}

	// 构建权限对象列表
	var permissions []models.RoleFieldPermission
	for _, fp := range req.Permissions {
		permissions = append(permissions, models.RoleFieldPermission{
			TenantID:       tenantIDUint64,
			RoleID:         roleID,
			EntityTable:    tableName,
			FieldName:      fp.FieldName,
			PermissionType: fp.PermissionType,
		})
	}

	// 更新角色字段权限
	err = h.fieldPermissionService.UpdateRoleFieldPermissions(ctx, roleID, tableName, permissions)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to update role field permissions",
			zap.Error(err),
			zap.Uint64("role_id", roleID),
			zap.String("table_name", tableName))
		h.responseWriter.Error(c, err)
		return
	}

	h.logger.InfoWithTrace(ctx, "Role field permissions updated successfully",
		zap.Uint64("role_id", roleID),
		zap.String("table_name", tableName),
		zap.Int("permission_count", len(permissions)))

	h.responseWriter.Success(c, dto.UpdateFieldPermissionsResponse{
		Message:         "字段权限更新成功",
		RoleID:          roleID,
		TableName:       tableName,
		PermissionCount: len(permissions),
	})
}

// GetUserFieldPermissions 获取用户的字段权限
// 注意：此方法未在路由中使用，实际API由UserHandler.GetUserFieldPermissions处理
func (h *FieldPermissionHandler) GetUserFieldPermissions(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取表名
	tableName := c.Param("tableName")
	if tableName == "" {
		h.logger.WarnWithTrace(ctx, "Missing table name parameter")
		h.responseWriter.BadRequest(c, "Table name is required")
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.WarnWithTrace(ctx, "User ID not found in context")
		h.responseWriter.Error(c, errors.ErrUnauthorized())
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		h.logger.WarnWithTrace(ctx, "Invalid user ID type")
		h.responseWriter.Error(c, errors.ErrInternalError("invalid user ID"))
		return
	}

	// 获取租户ID
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		h.logger.WarnWithTrace(ctx, "Tenant ID not found in context")
		h.responseWriter.Error(c, errors.ErrInternalError("tenant context not found"))
		return
	}

	tenantIDStr, ok := tenantID.(string)
	if !ok {
		h.logger.WarnWithTrace(ctx, "Invalid tenant ID type")
		h.responseWriter.Error(c, errors.ErrInternalError("invalid tenant context"))
		return
	}

	// 获取用户字段权限
	permissions, err := h.fieldPermissionService.GetUserFieldPermissions(ctx, userIDStr, tenantIDStr, tableName)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to get user field permissions",
			zap.Error(err),
			zap.String("user_id", userIDStr),
			zap.String("table_name", tableName))
		h.responseWriter.Error(c, err)
		return
	}

	// 转换为interface{}数组
	interfacePermissions := make([]interface{}, 0, len(permissions))
	for k, v := range permissions {
		interfacePermissions = append(interfacePermissions, map[string]interface{}{
			"field_name": k,
			"permission": v,
		})
	}

	h.responseWriter.Success(c, dto.UserFieldPermissionsResponse{
		TableName: tableName,
		Fields:    interfacePermissions,
	})
}

// InitializeTableFields 初始化表的字段权限配置
// @Summary 初始化表字段权限
// @Description 为指定表初始化字段权限配置，需要系统管理员权限
// @Tags system
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Param fields body dto.InitializeFieldsRequest true "字段配置列表"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Security BearerAuth
// @Router /system/field-permissions/{tableName}/initialize [post]
func (h *FieldPermissionHandler) InitializeTableFields(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取表名
	tableName := c.Param("tableName")
	if tableName == "" {
		h.logger.WarnWithTrace(ctx, "Missing table name parameter")
		h.responseWriter.BadRequest(c, "Table name is required")
		return
	}

	// 解析请求体
	var req dto.InitializeFieldsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid request body for initialize table fields")
		h.responseWriter.ValidationError(c, err)
		return
	}

	// 初始化字段权限配置
	err := h.fieldPermissionService.InitializeFieldPermissions(ctx, tableName, req.Fields)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to initialize field permissions",
			zap.Error(err),
			zap.String("table_name", tableName))
		h.responseWriter.Error(c, err)
		return
	}

	h.logger.InfoWithTrace(ctx, "Field permissions initialized successfully",
		zap.String("table_name", tableName),
		zap.Int("field_count", len(req.Fields)))

	h.responseWriter.Success(c, dto.InitializeFieldsResponse{
		Message:    "字段权限配置初始化成功",
		TableName:  tableName,
		FieldCount: len(req.Fields),
	})
}
