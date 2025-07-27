package dto

// UpdateRoleFieldPermissionsRequest 更新角色字段权限请求
type UpdateRoleFieldPermissionsRequest struct {
	FieldPermissions []FieldPermissionConfig `json:"field_permissions" binding:"required"`
}

// FieldPermissionConfig 字段权限配置
type FieldPermissionConfig struct {
	FieldName      string `json:"field_name" binding:"required"`
	PermissionType string `json:"permission_type" binding:"required"`
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