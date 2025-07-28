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

// RoleHandler 角色处理器
type RoleHandler struct {
	roleService    services.RoleService
	logger         *logger.Logger
	responseWriter *response.ResponseWriter
}

// NewRoleHandler 创建角色处理器
func NewRoleHandler(
	roleService services.RoleService,
	logger *logger.Logger,
) *RoleHandler {
	return &RoleHandler{
		roleService:    roleService,
		logger:         logger,
		responseWriter: response.NewResponseWriter(logger),
	}
}

// CreateRole 创建角色
// @Summary 创建角色
// @Description 创建新的角色，需要租户管理员权限
// @Tags roles
// @Accept json
// @Produce json
// @Param role body dto.CreateRoleRequest true "角色信息"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Security BearerAuth
// @Router /roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.CreateRoleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid request body for create role")
		h.responseWriter.ValidationError(c, err)
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

	// 转换tenantID为uint64
	tenantIDUint64, err := strconv.ParseUint(tenantIDStr, 10, 64)
	if err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid tenant ID format",
			zap.String("tenant_id", tenantIDStr))
		h.responseWriter.Error(c, errors.ErrInternalError("invalid tenant ID"))
		return
	}

	// 创建角色对象
	role := &models.Role{
		TenantModel: models.TenantModel{TenantID: tenantIDUint64},
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Type:        models.RoleTypeCustom,
		IsActive:    true,
	}

	createdRole, err := h.roleService.CreateRole(ctx, role)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to create role",
			zap.Error(err),
			zap.String("code", req.Code))
		h.responseWriter.Error(c, err)
		return
	}

	h.logger.InfoWithTrace(ctx, "Role created successfully",
		zap.String("code", req.Code))

	h.responseWriter.Created(c, dto.RoleResponse{
		ID:          createdRole.ID,
		UUID:        createdRole.UUID,
		Code:        createdRole.Code,
		Name:        createdRole.Name,
		Description: createdRole.Description,
		IsActive:    createdRole.IsActive,
		TenantID:    createdRole.TenantID,
	})
}

// ListRoles 获取角色列表
// @Summary 获取角色列表
// @Description 获取租户内的角色列表，支持分页
// @Tags roles
// @Accept json
// @Produce json
// @Param page query int false "页码，默认1"
// @Param limit query int false "每页数量，默认20"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Security BearerAuth
// @Router /roles [get]
func (h *RoleHandler) ListRoles(c *gin.Context) {
	ctx := c.Request.Context()

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
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

	roles, total, err := h.roleService.ListRoles(ctx, tenantIDStr, page, limit)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to list roles",
			zap.Error(err),
			zap.String("tenant_id", tenantIDStr))
		h.responseWriter.Error(c, err)
		return
	}

	// 转换为响应DTO
	roleResponses := make([]dto.RoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = dto.RoleResponse{
			ID:          role.ID,
			UUID:        role.UUID,
			Code:        role.Code,
			Name:        role.Name,
			Description: role.Description,
			IsActive:    role.IsActive,
			TenantID:    role.TenantID,
		}
	}

	h.responseWriter.Success(c, dto.RoleListResponse{
		Roles: roleResponses,
		Pagination: dto.PaginationMeta{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	})
}

// GetRole 获取单个角色
// @Summary 获取角色信息
// @Description 根据ID获取角色详细信息
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /roles/{id} [get]
func (h *RoleHandler) GetRole(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取角色ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid role ID",
			zap.String("id", idStr))
		h.responseWriter.BadRequest(c, "Invalid role ID")
		return
	}

	// TODO: 暂时返回模拟数据，等待Service方法实现
	// role, err := h.roleService.GetRoleByID(ctx, id)
	// if err != nil {
	// 	h.logger.ErrorWithTrace(ctx, "Failed to get role",
	// 		zap.Uint64("role_id", id),
	// 		zap.Error(err))
	// 	h.responseWriter.Error(c, err)
	// 	return
	// }

	role := dto.RoleResponse{
		ID:          id,
		UUID:        "sample-uuid",
		Code:        "sample_role",
		Name:        "示例角色",
		Description: "这是一个示例角色",
		IsActive:    true,
		TenantID:    1,
	}

	h.responseWriter.Success(c, role)
}

// UpdateRole 更新角色
// @Summary 更新角色
// @Description 更新角色信息，需要租户管理员权限
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param role body dto.UpdateRoleRequest true "角色更新信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取角色ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid role ID",
			zap.String("id", idStr))
		h.responseWriter.BadRequest(c, "Invalid role ID")
		return
	}

	var req dto.UpdateRoleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid request body for update role")
		h.responseWriter.ValidationError(c, err)
		return
	}

	// TODO: 暂时返回模拟数据，等待Service方法实现
	// 这里应该调用 h.roleService.UpdateRole(ctx, role)

	updatedRole := dto.RoleResponse{
		ID:          id,
		UUID:        "sample-uuid",
		Code:        "sample_role",
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive != nil && *req.IsActive,
		TenantID:    1,
	}

	h.logger.InfoWithTrace(ctx, "Role updated successfully (mocked)",
		zap.Uint64("role_id", id))

	h.responseWriter.Success(c, updatedRole)
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 删除指定角色，需要租户管理员权限
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取角色ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid role ID",
			zap.String("id", idStr))
		h.responseWriter.BadRequest(c, "Invalid role ID")
		return
	}

	// TODO: 暂时模拟删除，等待Service方法实现
	// err := h.roleService.DeleteRole(ctx, id)
	// if err != nil {
	// 	h.logger.ErrorWithTrace(ctx, "Failed to delete role",
	// 		zap.Uint64("role_id", id),
	// 		zap.Error(err))
	// 	h.responseWriter.Error(c, err)
	// 	return
	// }

	h.logger.InfoWithTrace(ctx, "Role deleted successfully (mocked)",
		zap.Uint64("role_id", id))

	h.responseWriter.NoContent(c)
}

// AssignPermissions 分配权限给角色
// @Summary 分配权限给角色
// @Description 为角色分配权限列表，需要租户管理员权限
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param permissions body dto.AssignPermissionsRequest true "权限ID列表"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /roles/{id}/permissions [post]
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取角色ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid role ID",
			zap.String("id", idStr))
		h.responseWriter.BadRequest(c, "Invalid role ID")
		return
	}

	var req dto.AssignPermissionsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid request body for assign permissions")
		h.responseWriter.ValidationError(c, err)
		return
	}

	// 分配权限给角色
	err = h.roleService.AssignPermissions(ctx, id, req.PermissionIDs)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to assign permissions to role",
			zap.Uint64("role_id", id),
			zap.Error(err))
		h.responseWriter.Error(c, err)
		return
	}

	h.logger.InfoWithTrace(ctx, "Permissions assigned to role successfully",
		zap.Uint64("role_id", id),
		zap.Int("permission_count", len(req.PermissionIDs)))

	h.responseWriter.Success(c, dto.AssignPermissionsResponse{
		Message:         "权限分配成功",
		RoleID:          id,
		PermissionCount: len(req.PermissionIDs),
	})
}

// GetRolePermissions 获取角色权限
// @Summary 获取角色权限
// @Description 获取指定角色的权限列表
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /roles/{id}/permissions [get]
func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取角色ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid role ID",
			zap.String("id", idStr))
		h.responseWriter.BadRequest(c, "Invalid role ID")
		return
	}

	// 获取角色权限
	permissions, err := h.roleService.GetRolePermissions(ctx, id)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to get role permissions",
			zap.Uint64("role_id", id),
			zap.Error(err))
		h.responseWriter.Error(c, err)
		return
	}

	// 转换为interface{}数组
	interfacePermissions := make([]interface{}, len(permissions))
	for i, perm := range permissions {
		interfacePermissions[i] = perm
	}

	h.responseWriter.Success(c, dto.RolePermissionsResponse{
		RoleID:      id,
		Permissions: interfacePermissions,
	})
}
