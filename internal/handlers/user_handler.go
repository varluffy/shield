// Package handlers contains HTTP request handlers for the API endpoints.
// It processes HTTP requests and returns appropriate responses.
package handlers

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/varluffy/shield/internal/dto"
	"github.com/varluffy/shield/internal/services"
	"github.com/varluffy/shield/pkg/errors"
	"github.com/varluffy/shield/pkg/logger"
	"github.com/varluffy/shield/pkg/response"
	"go.uber.org/zap"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService       services.UserService
	permissionService services.PermissionService
	logger            *logger.Logger
	responseWriter    *response.ResponseWriter
}

// NewUserHandler 创建用户处理器
func NewUserHandler(
	userService services.UserService,
	permissionService services.PermissionService,
	logger *logger.Logger,
) *UserHandler {
	return &UserHandler{
		userService:       userService,
		permissionService: permissionService,
		logger:            logger,
		responseWriter:    response.NewResponseWriter(logger),
	}
}

// CreateUser 创建用户（管理员权限）
// @Summary 创建用户账号（管理员权限）
// @Description 创建新的用户账号，需要管理员权限
// @Tags admin
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "用户信息"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Security BearerAuth
// @Router /admin/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest

	// 绑定和验证参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(c.Request.Context(), "Invalid request body for create user")
		h.responseWriter.ValidationError(c, err)
		return
	}

	// 从Gin上下文中获取tenant_id并添加到request context中
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		h.logger.WarnWithTrace(c.Request.Context(), "Tenant ID not found in context")
		h.responseWriter.Error(c, errors.ErrInternalError("tenant context not found"))
		return
	}

	// 现在tenant_id应该是uint64类型
	var tenantIDUint64 uint64
	switch v := tenantID.(type) {
	case uint64:
		tenantIDUint64 = v
	case string:
		// 兼容性处理：如果仍然是字符串，尝试转换
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			h.logger.WarnWithTrace(c.Request.Context(), "Invalid tenant ID format", zap.String("tenant_id", v))
			h.responseWriter.Error(c, errors.ErrInternalError("invalid tenant context"))
			return
		}
		tenantIDUint64 = parsed
	default:
		h.logger.WarnWithTrace(c.Request.Context(), "Invalid tenant ID type")
		h.responseWriter.Error(c, errors.ErrInternalError("invalid tenant context"))
		return
	}

	// 创建包含tenant_id的context
	ctx := context.WithValue(c.Request.Context(), "tenant_id", tenantIDUint64)

	user, err := h.userService.CreateUser(ctx, req)
	if err != nil {
		h.logger.WarnWithTrace(c.Request.Context(), "Failed to create user",
			zap.Error(err),
		)
		h.responseWriter.Error(c, err)
		return
	}

	h.logger.InfoWithTrace(c.Request.Context(), "User created successfully",
		zap.String("user_uuid", user.ID), // 现在返回的是UUID
	)

	h.responseWriter.Created(c, user)
}

// GetUser 获取单个用户信息
// @Summary 获取用户信息
// @Description 获取用户详细信息，用户只能查看自己的信息，管理员可以查看任意用户
// @Tags users
// @Accept json
// @Produce json
// @Param uuid path string true "用户UUID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /users/{uuid} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		h.logger.WarnWithTrace(c.Request.Context(), "Missing user UUID parameter")
		h.responseWriter.BadRequest(c, "User UUID is required")
		return
	}

	user, err := h.userService.GetUserByUUID(c.Request.Context(), uuid)
	if err != nil {
		h.logger.WarnWithTrace(c.Request.Context(), "Failed to get user",
			zap.String("user_uuid", uuid),
			zap.Error(err),
		)
		h.responseWriter.Error(c, err)
		return
	}

	h.responseWriter.Success(c, user)
}

// UpdateUser 更新用户
// @Summary 更新用户信息
// @Description 更新用户信息，用户只能更新自己的信息，管理员可以更新任意用户
// @Tags users
// @Accept json
// @Produce json
// @Param uuid path string true "用户UUID"
// @Param user body dto.UpdateUserRequest true "更新信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /users/{uuid} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		h.logger.WarnWithTrace(c.Request.Context(), "Missing user UUID parameter")
		h.responseWriter.BadRequest(c, "User UUID is required")
		return
	}

	var req dto.UpdateUserRequest

	// 绑定和验证参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(c.Request.Context(), "Invalid request body for update user")
		h.responseWriter.ValidationError(c, err)
		return
	}

	user, err := h.userService.UpdateUserByUUID(c.Request.Context(), uuid, req)
	if err != nil {
		h.logger.WarnWithTrace(c.Request.Context(), "Failed to update user",
			zap.String("user_uuid", uuid),
			zap.Error(err),
		)
		h.responseWriter.Error(c, err)
		return
	}

	h.logger.InfoWithTrace(c.Request.Context(), "User updated successfully",
		zap.String("user_uuid", user.ID), // 现在返回的是UUID
	)

	h.responseWriter.Success(c, user)
}

// DeleteUser 删除用户
// @Summary 删除用户（管理员权限）
// @Description 删除指定用户，只有管理员可以执行此操作
// @Tags users
// @Accept json
// @Produce json
// @Param uuid path string true "用户UUID"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /users/{uuid} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		h.logger.WarnWithTrace(c.Request.Context(), "Missing user UUID parameter")
		h.responseWriter.BadRequest(c, "User UUID is required")
		return
	}

	if err := h.userService.DeleteUserByUUID(c.Request.Context(), uuid); err != nil {
		h.logger.WarnWithTrace(c.Request.Context(), "Failed to delete user",
			zap.String("user_uuid", uuid),
			zap.Error(err),
		)
		h.responseWriter.Error(c, err)
		return
	}

	h.logger.InfoWithTrace(c.Request.Context(), "User deleted successfully",
		zap.String("user_uuid", uuid),
	)

	h.responseWriter.NoContent(c)
}

// ListUsers 获取用户列表
// @Summary 获取用户列表（管理员权限）
// @Description 获取系统中所有用户的列表，支持分页和筛选，需要管理员权限
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Param name query string false "用户名筛选"
// @Param email query string false "邮箱筛选"
// @Param role query string false "角色筛选"
// @Param active query bool false "激活状态筛选"
// @Param order_by query string false "排序字段" default(created_at)
// @Param order_dir query string false "排序方向" default(DESC)
// @Success 200 {object} response.PaginationResponse
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Failure 403 {object} response.Response "权限不足"
// @Security BearerAuth
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 构建筛选条件
	filter := dto.UserFilter{}

	// 解析分页参数
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}
	if filter.Page == 0 {
		filter.Page = 1
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if filter.Limit == 0 {
		filter.Limit = 10
	}

	// 解析筛选参数
	filter.Name = c.Query("name")
	filter.Email = c.Query("email")
	filter.Role = c.Query("role")

	if activeStr := c.Query("active"); activeStr != "" {
		if active, err := strconv.ParseBool(activeStr); err == nil {
			filter.Active = &active
		}
	}

	// 解析排序参数
	filter.OrderBy = c.DefaultQuery("order_by", "created_at")
	filter.OrderDir = c.DefaultQuery("order_dir", "DESC")

	// 获取用户列表
	result, err := h.userService.ListUsers(c.Request.Context(), filter)
	if err != nil {
		h.logger.ErrorWithTrace(c.Request.Context(), "Failed to list users",
			zap.Any("filter", filter),
			zap.Error(err),
		)
		h.responseWriter.Error(c, err)
		return
	}

	// 构建分页元数据
	meta := &response.PaginationMeta{
		Page:       filter.Page,
		Limit:      filter.Limit,
		Total:      int64(result.Meta.Total),
		TotalPages: result.Meta.TotalPage,
	}

	h.responseWriter.Pagination(c, result.Users, meta)
}

// Login 用户登录
// @Summary 用户登录
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequest true "登录凭据"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	// 绑定和验证参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(c.Request.Context(), "Invalid request body for login")
		h.responseWriter.ValidationError(c, err)
		return
	}

	user, err := h.userService.Login(c.Request.Context(), req)
	if err != nil {
		h.logger.WarnWithTrace(c.Request.Context(), "Failed to login",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		h.responseWriter.Error(c, err)
		return
	}

	h.logger.InfoWithTrace(c.Request.Context(), "User login successfully",
		zap.String("email", req.Email),
		zap.String("user_id", user.User.ID),
	)

	h.responseWriter.Success(c, user)
}


// Register 用户注册
// @Summary 用户注册
// @Tags auth
// @Accept json
// @Produce json
// @Param registration body dto.RegisterRequest true "注册信息"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	// 绑定和验证参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(c.Request.Context(), "Invalid request body for register")
		h.responseWriter.ValidationError(c, err)
		return
	}

	user, err := h.userService.Register(c.Request.Context(), req)
	if err != nil {
		h.logger.WarnWithTrace(c.Request.Context(), "Failed to register user",
			zap.String("email", req.Email),
			zap.Error(err),
		)
		h.responseWriter.Error(c, err)
		return
	}

	h.logger.InfoWithTrace(c.Request.Context(), "User registered successfully",
		zap.String("email", req.Email),
		zap.String("user_id", user.ID),
	)

	h.responseWriter.Created(c, user)
}

// RefreshToken 刷新访问令牌
// @Summary 刷新访问令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body dto.RefreshTokenRequest true "刷新令牌"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /auth/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest

	// 绑定和验证参数
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnWithTrace(c.Request.Context(), "Invalid request body for refresh token")
		h.responseWriter.ValidationError(c, err)
		return
	}

	response, err := h.userService.RefreshToken(c.Request.Context(), req)
	if err != nil {
		h.logger.WarnWithTrace(c.Request.Context(), "Failed to refresh token",
			zap.Error(err),
		)
		h.responseWriter.Error(c, err)
		return
	}

	h.logger.InfoWithTrace(c.Request.Context(), "Token refreshed successfully")
	h.responseWriter.Success(c, response)
}

// GetUserPermissions 获取当前用户权限列表
// @Summary 获取当前用户权限列表
// @Description 获取当前登录用户的所有权限信息
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Security BearerAuth
// @Router /user/permissions [get]
func (h *UserHandler) GetUserPermissions(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	if userID == "" || tenantID == "" {
		h.logger.WarnWithTrace(ctx, "Missing user_id or tenant_id in context")
		h.responseWriter.Error(c, errors.ErrUnauthorized())
		return
	}

	// 这里需要调用权限服务获取用户权限
	// 暂时返回空数组，实际应该调用 permissionService.GetUserPermissions
	permissions := dto.UserPermissionsResponse{
		Menus:   []string{},
		Buttons: []string{},
		APIs:    []string{},
	}

	h.responseWriter.Success(c, permissions)
}

// GetUserMenuPermissions 获取当前用户菜单权限
// @Summary 获取当前用户菜单权限
// @Description 获取当前登录用户的菜单权限树
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Security BearerAuth
// @Router /user/permissions/menu [get]
func (h *UserHandler) GetUserMenuPermissions(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	if userID == "" || tenantID == "" {
		h.logger.WarnWithTrace(ctx, "Missing user_id or tenant_id in context")
		h.responseWriter.Error(c, errors.ErrUnauthorized())
		return
	}

	h.logger.DebugWithTrace(ctx, "Getting user menu permissions",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID))

	// 获取用户菜单权限
	menuPermissions, err := h.permissionService.GetUserMenuPermissions(ctx, userID, tenantID)
	if err != nil {
		h.logger.ErrorWithTrace(ctx, "Failed to get user menu permissions",
			zap.String("user_id", userID),
			zap.String("tenant_id", tenantID),
			zap.Error(err))
		h.responseWriter.Error(c, errors.ErrInternalError("failed to get menu permissions"))
		return
	}

	// 构建菜单树
	menuTree := h.permissionService.BuildMenuTree(ctx, menuPermissions)

	// 转换为DTO格式
	response := h.buildMenuResponse(menuTree)

	h.logger.InfoWithTrace(ctx, "Successfully retrieved user menu permissions",
		zap.String("user_id", userID),
		zap.String("tenant_id", tenantID),
		zap.Int("menu_count", len(response.Menus)))

	h.responseWriter.Success(c, response)
}

// buildMenuResponse 构建菜单响应
func (h *UserHandler) buildMenuResponse(menuTree []interface{}) dto.UserMenuPermissionsResponse {
	var menus []dto.MenuItemResponse

	for _, item := range menuTree {
		if menuItem, ok := item.(map[string]interface{}); ok {
			menu := h.convertToMenuDTO(menuItem)
			menus = append(menus, menu)
		}
	}

	return dto.UserMenuPermissionsResponse{
		Menus: menus,
	}
}

// convertToMenuDTO 转换为菜单DTO
func (h *UserHandler) convertToMenuDTO(item map[string]interface{}) dto.MenuItemResponse {
	menu := dto.MenuItemResponse{
		ID:   getString(item, "id"),
		Name: getString(item, "name"),
		Icon: getString(item, "icon"),
		Path: getString(item, "path"),
		Sort: getInt(item, "sort"),
		Type: getString(item, "type"),
	}

	// 处理子菜单
	if children, ok := item["children"].([]interface{}); ok {
		for _, child := range children {
			if childItem, ok := child.(map[string]interface{}); ok {
				childMenu := h.convertToMenuDTO(childItem)
				menu.Children = append(menu.Children, childMenu)
			}
		}
	}

	return menu
}

// getString 安全获取字符串值
func getString(item map[string]interface{}, key string) string {
	if val, ok := item[key].(string); ok {
		return val
	}
	return ""
}

// getInt 安全获取整数值
func getInt(item map[string]interface{}, key string) int {
	if val, ok := item[key].(int); ok {
		return val
	}
	// 尝试从其他数值类型转换
	if val, ok := item[key].(float64); ok {
		return int(val)
	}
	if val, ok := item[key].(int64); ok {
		return int(val)
	}
	return 0
}

// GetUserFieldPermissions 获取当前用户字段权限
// @Summary 获取当前用户字段权限
// @Description 获取当前登录用户对指定表的字段权限
// @Tags user
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response "未授权"
// @Security BearerAuth
// @Router /user/field-permissions/{tableName} [get]
func (h *UserHandler) GetUserFieldPermissions(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")
	tableName := c.Param("tableName")

	if userID == "" || tenantID == "" {
		h.logger.WarnWithTrace(ctx, "Missing user_id or tenant_id in context")
		h.responseWriter.Error(c, errors.ErrUnauthorized())
		return
	}

	if tableName == "" {
		h.logger.WarnWithTrace(ctx, "Missing table name parameter")
		h.responseWriter.Error(c, errors.ErrInvalidRequest())
		return
	}

	// 这里需要调用字段权限服务获取用户字段权限
	// 暂时返回空对象，实际应该调用 fieldPermissionService.GetUserFieldPermissions
	fieldPermissions := dto.UserFieldPermissionsResponse{
		TableName: tableName,
		Fields:    []interface{}{},
	}

	h.responseWriter.Success(c, fieldPermissions)
}
