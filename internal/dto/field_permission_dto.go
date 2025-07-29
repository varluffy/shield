package dto

import "time"

// FieldMetadataResponse 字段元数据响应
type FieldMetadataResponse struct {
	Tables          []TableFields       `json:"tables"`
	PermissionTypes []PermissionTypeInfo `json:"permission_types"`
}

// TableMetadata 表元数据
type TableMetadata struct {
	TableName   string `json:"table_name"`
	TableLabel  string `json:"table_label"`
	Description string `json:"description"`
}

// TableFields 表字段信息
type TableFields struct {
	TableMetadata
	Fields []FieldMetadata `json:"fields"`
}

// FieldMetadata 字段元数据
type FieldMetadata struct {
	FieldName    string `json:"field_name"`
	FieldLabel   string `json:"field_label"`
	FieldType    string `json:"field_type"`
	Description  string `json:"description"`
	DefaultValue string `json:"default_value"`
	SortOrder    int    `json:"sort_order"`
	IsActive     bool   `json:"is_active"`
}

// PermissionTypeInfo 权限类型信息
type PermissionTypeInfo struct {
	Type        string `json:"type"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// TableFieldsResponse 表字段响应
type TableFieldsResponse struct {
	TableName string          `json:"table_name"`
	Fields    []FieldMetadata `json:"fields"`
}

// RoleFieldPermissionsResponse 角色字段权限响应
type RoleFieldPermissionsResponse struct {
	RoleID           uint64                   `json:"role_id"`
	TableName        string                   `json:"table_name"`
	FieldPermissions []FieldPermissionConfig  `json:"field_permissions"`
	LastModified     *time.Time              `json:"last_modified"`
}

// UpdateRoleFieldPermissionsRequest 更新角色字段权限请求
type UpdateRoleFieldPermissionsRequest struct {
	Permissions []FieldPermissionUpdate `json:"permissions" binding:"required"`
}

// FieldPermissionUpdate 字段权限更新配置
type FieldPermissionUpdate struct {
	FieldName      string `json:"field_name" binding:"required"`
	PermissionType string `json:"permission_type" binding:"required"`
}

// FieldPermissionConfig 字段权限配置
type FieldPermissionConfig struct {
	FieldName      string `json:"field_name" binding:"required"`
	FieldLabel     string `json:"field_label"`
	FieldType      string `json:"field_type"`
	DefaultValue   string `json:"default_value"`
	PermissionType string `json:"permission_type" binding:"required"`
	Description    string `json:"description"`
	SortOrder      int    `json:"sort_order"`
}

// InitializeFieldsRequest 初始化字段配置请求
type InitializeFieldsRequest struct {
	Fields []FieldConfig `json:"fields" binding:"required"`
}

// FieldConfig 字段配置
type FieldConfig struct {
	FieldName    string `json:"field_name" binding:"required"`
	FieldLabel   string `json:"field_label"`
	FieldType    string `json:"field_type" binding:"required"`
	DefaultValue string `json:"default_value"`
	Description  string `json:"description"`
	SortOrder    int    `json:"sort_order"`
	IsRequired   bool   `json:"is_required"`
}

// UpdateFieldPermissionsResponse 更新字段权限响应
type UpdateFieldPermissionsResponse struct {
	Message         string `json:"message"`
	RoleID          uint64 `json:"role_id"`
	TableName       string `json:"table_name"`
	PermissionCount int    `json:"permission_count"`
}

// UserFieldPermissionsResponse 用户字段权限响应
type UserFieldPermissionsResponse struct {
	TableName string        `json:"table_name"`
	Fields    []interface{} `json:"fields"`
}

// InitializeFieldsResponse 初始化字段权限响应
type InitializeFieldsResponse struct {
	Message    string `json:"message"`
	TableName  string `json:"table_name"`
	FieldCount int    `json:"field_count"`
}
