// Package handlers contains HTTP request handlers for the API endpoints.
package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/varluffy/shield/internal/dto"
	"github.com/varluffy/shield/internal/services"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/response"
	"go.uber.org/zap"
)

// PermissionHandler 权限处理器
type PermissionHandler struct {
	permissionService services.PermissionService
	logger            *logger.Logger
	responseWriter    *response.ResponseWriter
}

// NewPermissionHandler 创建权限处理器
func NewPermissionHandler(
	permissionService services.PermissionService,
	logger *logger.Logger,
) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
		logger:            logger,
		responseWriter:    response.NewResponseWriter(logger),
	}
}

// ListPermissions 获取权限列表（自动根据用户身份过滤）
// @Summary 获取权限列表
// @Description 根据用户身份自动返回对应的权限列表：系统管理员看到所有权限，租户管理员只看到租户权限
// @Tags permissions
// @Accept json
// @Produce json
// @Param module query string false "权限模块：user, role, system等"
// @Param type query string false "权限类型：menu, button, api"
// @Param page query int false "页码，默认1"
// @Param limit query int false "每页数量，默认20"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Security BearerAuth
// @Router /permissions [get]
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取查询参数（移除scope参数，由后端自动判断）
	module := c.Query("module")
	permType := c.Query("type")
	
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// 构造筛选条件
	filter := map[string]interface{}{}
	
	if module != "" {
		filter["module"] = module
	}
	if permType != "" {
		filter["type"] = permType
	}

	// 调用服务获取权限列表（服务层会自动根据用户身份过滤）
	permissions, total, err := h.permissionService.ListPermissions(ctx, filter, page, limit)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to list permissions",
			zap.Error(err))
		h.responseWriter.Error(c, err)
		return
	}

	// 转换为interface{}数组
	interfacePermissions := make([]interface{}, len(permissions))
	for i, perm := range permissions {
		interfacePermissions[i] = perm
	}

	h.responseWriter.Success(c, dto.PermissionListResponse{
		Permissions: interfacePermissions,
		Pagination: dto.PaginationMeta{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	})
}

// GetPermissionTree 获取权限树结构（自动根据用户身份过滤）
// @Summary 获取权限树结构
// @Description 根据用户身份自动返回权限树：系统管理员可指定scope或获取全部，租户管理员只能获取租户权限树
// @Tags permissions
// @Accept json
// @Produce json
// @Param scope query string false "权限作用域：system, tenant（可选，系统管理员可用）"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Security BearerAuth
// @Router /permissions/tree [get]
func (h *PermissionHandler) GetPermissionTree(c *gin.Context) {
	ctx := c.Request.Context()

	// scope参数变为可选，由后端根据用户身份自动处理
	scope := c.Query("scope")

	// 获取租户ID（主要用于日志记录）
	tenantID, exists := c.Get("tenant_id")
	tenantIDStr := "unknown"
	if exists {
		if tid, ok := tenantID.(string); ok {
			tenantIDStr = tid
		}
	}

	// 调用服务获取权限树（服务层会自动根据用户身份过滤）
	tree, err := h.permissionService.GetPermissionTree(ctx, tenantIDStr, scope)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to get permission tree",
			zap.Error(err),
			zap.String("tenant_id", tenantIDStr),
			zap.String("scope", scope))
		h.responseWriter.Error(c, err)
		return
	}

	// 转换tree为数组格式
	var treeArray []interface{}
	if tree != nil {
		if arr, ok := tree.([]interface{}); ok {
			treeArray = arr
		} else {
			treeArray = []interface{}{tree}
		}
	}

	h.responseWriter.Success(c, dto.PermissionTreeResponse{
		PermissionTree: treeArray,
	})
}

// ListSystemPermissions 获取系统权限列表（已废弃，建议使用 /permissions 接口）
// @Summary 获取系统权限列表（已废弃）
// @Description ⚠️ 已废弃：建议使用 GET /permissions 接口，后端会自动根据用户身份过滤权限
// @Tags system
// @Deprecated true
// @Accept json
// @Produce json
// @Param page query int false "页码，默认1"
// @Param limit query int false "每页数量，默认20"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Security BearerAuth
// @Router /system/permissions [get]
func (h *PermissionHandler) ListSystemPermissions(c *gin.Context) {
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

	// 调用服务获取系统权限列表
	permissions, total, err := h.permissionService.ListSystemPermissions(ctx, page, limit)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to list system permissions",
			zap.Error(err))
		h.responseWriter.Error(c, err)
		return
	}

	// 转换为interface{}数组
	interfacePermissions := make([]interface{}, len(permissions))
	for i, perm := range permissions {
		interfacePermissions[i] = perm
	}

	h.responseWriter.Success(c, dto.PermissionListResponse{
		Permissions: interfacePermissions,
		Pagination: dto.PaginationMeta{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	})
}

// UpdatePermission 更新权限（系统管理员专用）
// @Summary 更新权限
// @Description 更新系统权限信息，只有系统管理员可以访问
// @Tags system
// @Accept json
// @Produce json
// @Param id path int true "权限ID"
// @Param permission body dto.UpdatePermissionRequest true "权限更新信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /system/permissions/{id} [put]
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取权限ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid permission ID",
			zap.String("id", idStr))
		h.responseWriter.BadRequest(c, "Invalid permission ID")
		return
	}

	var req dto.UpdatePermissionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(ctx, "Invalid request body for update permission")
		h.responseWriter.ValidationError(c, err)
		return
	}

	permission, err := h.permissionService.UpdatePermission(ctx, id, req.Name, req.Description, req.IsActive)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to update permission",
			zap.Uint64("id", id),
			zap.Error(err))
		h.responseWriter.Error(c, err)
		return
	}

	h.logger.InfoWithTrace(ctx, "Permission updated successfully",
		zap.Uint64("id", id))

	h.responseWriter.Success(c, permission)
} 