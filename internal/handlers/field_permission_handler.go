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
		h.responseWriter.BadRequest(c, "Table name is required")
		return
	}

	fields, err := h.fieldPermissionService.GetTableFields(ctx, tableName)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to get table fields",
			zap.Error(err),
			zap.String("table_name", tableName))
		h.responseWriter.Error(c, err)
		return
	}

	// 转换为interface{}数组
	interfaceFields := make([]interface{}, len(fields))
	for i, field := range fields {
		interfaceFields[i] = field
	}

	h.responseWriter.Success(c, dto.TableFieldsResponse{
		TableName: tableName,
		Fields:    interfaceFields,
	})
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
	var fieldPermissions []interface{}
	for _, field := range fields {
		permission := field.DefaultValue // 使用默认权限
		if p, exists := permissionMap[field.FieldName]; exists {
			permission = p // 使用角色设置的权限
		}

		fieldPermissions = append(fieldPermissions, map[string]interface{}{
			"field_name":         field.FieldName,
			"field_label":        field.FieldLabel,
			"field_type":         field.FieldType,
			"default_value":      field.DefaultValue,
			"current_permission": permission,
			"description":        field.Description,
			"sort_order":         field.SortOrder,
			"is_active":          field.IsActive,
		})
	}

	h.responseWriter.Success(c, dto.RoleFieldPermissionsResponse{
		RoleID:           roleID,
		TableName:        tableName,
		FieldPermissions: fieldPermissions,
	})
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
	for _, fp := range req.FieldPermissions {
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
		Message:           "字段权限更新成功",
		RoleID:            roleID,
		TableName:         tableName,
		PermissionCount:   len(permissions),
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