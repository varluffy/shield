// Package dto contains data transfer objects for the API layer.
// It defines request and response structures used for HTTP communication.
package dto

import "time"

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=50" label:"姓名"`
	Email    string `json:"email" binding:"required,email" label:"邮箱"`
	Password string `json:"password" binding:"required,min=8,max=128" label:"密码"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Name   string `json:"name" binding:"omitempty,min=2,max=50" label:"姓名"`
	Email  string `json:"email" binding:"omitempty,email" label:"邮箱"`
	Active *bool  `json:"active" label:"激活状态"`
}

// UserResponse 用户响应（对外只暴露UUID，不暴露内部ID）
type UserResponse struct {
	ID        string    `json:"id"` // 使用UUID作为对外ID
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	Active    bool      `json:"active"`
	TenantID  string    `json:"tenant_id"` // 使用UUID作为对外Tenant ID
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Meta  PaginationMeta `json:"meta"`
}

// PaginationMeta 分页元数据
type PaginationMeta struct {
	Page      int `json:"page"`
	Limit     int `json:"limit"`
	Total     int `json:"total"`
	TotalPage int `json:"total_page"`
}

// UserFilter 用户筛选条件
type UserFilter struct {
	Name     string
	Email    string
	Role     string
	Active   *bool
	Page     int
	Limit    int
	OrderBy  string
	OrderDir string
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email     string `json:"email" binding:"required,email" example:"test@example.com"`
	Password  string `json:"password" binding:"required" example:"password123"`
	CaptchaID string `json:"captcha_id" binding:"required" example:"abc123"`
	Answer    string `json:"answer" binding:"required" example:"1234"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Name      string `json:"name" binding:"required,min=2,max=50" label:"姓名"`
	Email     string `json:"email" binding:"required,email" label:"邮箱"`
	Password  string `json:"password" binding:"required,min=8,max=128" label:"密码"`
	CaptchaID string `json:"captcha_id" binding:"required" label:"验证码ID"`
	Answer    string `json:"answer" binding:"required" label:"验证码"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" label:"刷新令牌"`
}

// RefreshTokenResponse 刷新令牌响应
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Code        string `json:"code" binding:"required" example:"admin"`
	Name        string `json:"name" binding:"required" example:"管理员"`
	Description string `json:"description" example:"系统管理员角色"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"required" example:"管理员"`
	Description string `json:"description" example:"系统管理员角色"`
	IsActive    *bool  `json:"is_active" example:"true"`
}

// AssignPermissionsRequest 分配权限请求
type AssignPermissionsRequest struct {
	PermissionIDs []uint64 `json:"permission_ids" binding:"required"`
}

// RoleResponse 角色响应
type RoleResponse struct {
	ID          uint64 `json:"id"`
	UUID        string `json:"uuid"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	TenantID    uint64 `json:"tenant_id"`
}

// RoleListResponse 角色列表响应
type RoleListResponse struct {
	Roles      []RoleResponse `json:"roles"`
	Pagination PaginationMeta `json:"pagination"`
}

// AssignPermissionsResponse 分配权限响应
type AssignPermissionsResponse struct {
	Message         string `json:"message"`
	RoleID          uint64 `json:"role_id"`
	PermissionCount int    `json:"permission_count"`
}

// RolePermissionsResponse 角色权限响应
type RolePermissionsResponse struct {
	RoleID      uint64        `json:"role_id"`
	Permissions []interface{} `json:"permissions"`
}

// PermissionListResponse 权限列表响应
type PermissionListResponse struct {
	Permissions []interface{}  `json:"permissions"`
	Pagination  PaginationMeta `json:"pagination"`
}

// PermissionTreeResponse 权限树响应
type PermissionTreeResponse struct {
	PermissionTree []interface{} `json:"permission_tree"`
}

// UserPermissionsResponse 用户权限响应
type UserPermissionsResponse struct {
	Menus   []string `json:"menus"`
	Buttons []string `json:"buttons"`
	APIs    []string `json:"apis"`
}

// MenuItemResponse 菜单项响应
type MenuItemResponse struct {
	ID       string             `json:"id"`       // 菜单ID（权限代码）
	Name     string             `json:"name"`     // 菜单名称
	Icon     string             `json:"icon"`     // 菜单图标
	Path     string             `json:"path"`     // 菜单路径
	Sort     int                `json:"sort"`     // 排序顺序
	Type     string             `json:"type"`     // 菜单类型（menu）
	Children []MenuItemResponse `json:"children"` // 子菜单
}

// UserMenuPermissionsResponse 用户菜单权限响应
type UserMenuPermissionsResponse struct {
	Menus []MenuItemResponse `json:"menus"`
}

