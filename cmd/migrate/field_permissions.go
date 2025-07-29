// Package main provides field permissions initialization utilities.
// This file handles the initialization of default field permissions for tables.
package main

import (
	"fmt"
	"log"

	"github.com/varluffy/shield/internal/models"
	"gorm.io/gorm"
)

// TableFieldConfig 表字段配置
type TableFieldConfig struct {
	TableName string
	Fields    []FieldConfig
}

// FieldConfig 字段配置
type FieldConfig struct {
	FieldName       string
	Label           string
	DefaultPermission string // default, readonly, hidden
	Description     string
}

// getDefaultFieldPermissions 获取默认字段权限配置
// 定义各个表的字段权限配置，支持default/readonly/hidden三种权限类型
func getDefaultFieldPermissions() []TableFieldConfig {
	return []TableFieldConfig{
		{
			TableName: "users",
			Fields: []FieldConfig{
				{
					FieldName:         "id",
					Label:             "用户ID",
					DefaultPermission: "readonly",
					Description:       "用户主键ID，通常只读",
				},
				{
					FieldName:         "uuid",
					Label:             "用户UUID",
					DefaultPermission: "readonly",
					Description:       "用户唯一标识，通常只读",
				},
				{
					FieldName:         "name",
					Label:             "姓名",
					DefaultPermission: "default",
					Description:       "用户姓名，可编辑",
				},
				{
					FieldName:         "email",
					Label:             "邮箱",
					DefaultPermission: "default",
					Description:       "用户邮箱，可编辑",
				},
				{
					FieldName:         "password",
					Label:             "密码",
					DefaultPermission: "hidden",
					Description:       "用户密码，隐藏显示",
				},
				{
					FieldName:         "status",
					Label:             "状态",
					DefaultPermission: "default",
					Description:       "用户状态（active/inactive）",
				},
				{
					FieldName:         "tenant_id",
					Label:             "租户ID",
					DefaultPermission: "readonly",
					Description:       "所属租户ID，通常只读",
				},
				{
					FieldName:         "created_at",
					Label:             "创建时间",
					DefaultPermission: "readonly",
					Description:       "记录创建时间，只读",
				},
				{
					FieldName:         "updated_at",
					Label:             "更新时间",
					DefaultPermission: "readonly",
					Description:       "记录更新时间，只读",
				},
			},
		},
		{
			TableName: "roles",
			Fields: []FieldConfig{
				{
					FieldName:         "id",
					Label:             "角色ID",
					DefaultPermission: "readonly",
					Description:       "角色主键ID，通常只读",
				},
				{
					FieldName:         "uuid",
					Label:             "角色UUID",
					DefaultPermission: "readonly",
					Description:       "角色唯一标识，通常只读",
				},
				{
					FieldName:         "code",
					Label:             "角色代码",
					DefaultPermission: "default",
					Description:       "角色唯一代码，可编辑",
				},
				{
					FieldName:         "name",
					Label:             "角色名称",
					DefaultPermission: "default",
					Description:       "角色显示名称，可编辑",
				},
				{
					FieldName:         "description",
					Label:             "角色描述",
					DefaultPermission: "default",
					Description:       "角色详细描述，可编辑",
				},
				{
					FieldName:         "type",
					Label:             "角色类型",
					DefaultPermission: "readonly",
					Description:       "角色类型（system/custom），通常只读",
				},
				{
					FieldName:         "tenant_id",
					Label:             "租户ID",
					DefaultPermission: "readonly",
					Description:       "所属租户ID，通常只读",
				},
				{
					FieldName:         "is_active",
					Label:             "激活状态",
					DefaultPermission: "default",
					Description:       "角色是否激活，可编辑",
				},
				{
					FieldName:         "created_at",
					Label:             "创建时间",
					DefaultPermission: "readonly",
					Description:       "记录创建时间，只读",
				},
				{
					FieldName:         "updated_at",
					Label:             "更新时间",
					DefaultPermission: "readonly",
					Description:       "记录更新时间，只读",
				},
			},
		},
		{
			TableName: "permissions",
			Fields: []FieldConfig{
				{
					FieldName:         "id",
					Label:             "权限ID",
					DefaultPermission: "readonly",
					Description:       "权限主键ID，通常只读",
				},
				{
					FieldName:         "uuid",
					Label:             "权限UUID",
					DefaultPermission: "readonly",
					Description:       "权限唯一标识，通常只读",
				},
				{
					FieldName:         "code",
					Label:             "权限代码",
					DefaultPermission: "readonly",
					Description:       "权限唯一代码，系统定义，只读",
				},
				{
					FieldName:         "name",
					Label:             "权限名称",
					DefaultPermission: "default",
					Description:       "权限显示名称，可编辑",
				},
				{
					FieldName:         "description",
					Label:             "权限描述",
					DefaultPermission: "default",
					Description:       "权限详细描述，可编辑",
				},
				{
					FieldName:         "type",
					Label:             "权限类型",
					DefaultPermission: "readonly",
					Description:       "权限类型（menu/button/api），只读",
				},
				{
					FieldName:         "scope",
					Label:             "权限范围",
					DefaultPermission: "readonly",
					Description:       "权限范围（system/tenant），只读",
				},
				{
					FieldName:         "parent_code",
					Label:             "父权限代码",
					DefaultPermission: "readonly",
					Description:       "父权限代码，系统定义，只读",
				},
				{
					FieldName:         "resource_path",
					Label:             "资源路径",
					DefaultPermission: "readonly",
					Description:       "API资源路径，只读",
				},
				{
					FieldName:         "method",
					Label:             "HTTP方法",
					DefaultPermission: "readonly",
					Description:       "HTTP请求方法，只读",
				},
				{
					FieldName:         "is_builtin",
					Label:             "内置标识",
					DefaultPermission: "readonly",
					Description:       "是否为系统内置权限，只读",
				},
				{
					FieldName:         "is_active",
					Label:             "激活状态",
					DefaultPermission: "default",
					Description:       "权限是否激活，可编辑",
				},
				{
					FieldName:         "created_at",
					Label:             "创建时间",
					DefaultPermission: "readonly",
					Description:       "记录创建时间，只读",
				},
				{
					FieldName:         "updated_at",
					Label:             "更新时间",
					DefaultPermission: "readonly",
					Description:       "记录更新时间，只读",
				},
			},
		},
		{
			TableName: "tenants",
			Fields: []FieldConfig{
				{
					FieldName:         "id",
					Label:             "租户ID",
					DefaultPermission: "readonly",
					Description:       "租户主键ID，通常只读",
				},
				{
					FieldName:         "uuid",
					Label:             "租户UUID",
					DefaultPermission: "readonly",
					Description:       "租户唯一标识，通常只读",
				},
				{
					FieldName:         "name",
					Label:             "租户名称",
					DefaultPermission: "default",
					Description:       "租户显示名称，可编辑",
				},
				{
					FieldName:         "domain",
					Label:             "租户域名",
					DefaultPermission: "default",
					Description:       "租户专用域名，可编辑",
				},
				{
					FieldName:         "status",
					Label:             "租户状态",
					DefaultPermission: "default",
					Description:       "租户状态（active/suspended），可编辑",
				},
				{
					FieldName:         "plan",
					Label:             "订阅计划",
					DefaultPermission: "default",
					Description:       "租户订阅计划，可编辑",
				},
				{
					FieldName:         "max_users",
					Label:             "最大用户数",
					DefaultPermission: "default",
					Description:       "租户最大用户数限制，可编辑",
				},
				{
					FieldName:         "max_storage",
					Label:             "最大存储空间",
					DefaultPermission: "default",
					Description:       "租户最大存储空间，可编辑",
				},
				{
					FieldName:         "created_at",
					Label:             "创建时间",
					DefaultPermission: "readonly",
					Description:       "记录创建时间，只读",
				},
				{
					FieldName:         "updated_at",
					Label:             "更新时间",
					DefaultPermission: "readonly",
					Description:       "记录更新时间，只读",
				},
			},
		},
	}
}

// InitializeFieldPermissions 初始化字段权限配置
func (m *Migration) InitializeFieldPermissions() error {
	fmt.Println("🚀 Starting field permissions initialization...")

	tableConfigs := getDefaultFieldPermissions()

	// 验证配置数据
	if err := validateFieldPermissionConfigs(tableConfigs); err != nil {
		return fmt.Errorf("字段权限配置验证失败: %v", err)
	}

	// 在事务中执行初始化
	return m.db.Transaction(func(tx *gorm.DB) error {
		m.db = tx // 在事务中执行所有操作

		for _, tableConfig := range tableConfigs {
			fmt.Printf("📋 Initializing field permissions for table: %s\n", tableConfig.TableName)

			for _, fieldConfig := range tableConfig.Fields {
				if err := m.initializeFieldPermission(tableConfig.TableName, fieldConfig); err != nil {
					return fmt.Errorf("初始化字段权限失败 %s.%s: %v", 
						tableConfig.TableName, fieldConfig.FieldName, err)
				}
			}

			fmt.Printf("  ✅ Initialized %d field permissions for %s\n", len(tableConfig.Fields), tableConfig.TableName)
		}

		fmt.Println("🎉 Field permissions initialization completed successfully!")
		return nil
	})
}

// initializeFieldPermission 初始化单个字段权限
func (m *Migration) initializeFieldPermission(tableName string, fieldConfig FieldConfig) error {
	var existingPerm models.FieldPermission
	result := m.db.Where("entity_table = ? AND field_name = ?", tableName, fieldConfig.FieldName).First(&existingPerm)

	if result.Error == gorm.ErrRecordNotFound {
		// 字段权限不存在，创建新配置
		fieldPerm := models.FieldPermission{
			EntityTable:  tableName,
			FieldName:    fieldConfig.FieldName,
			FieldLabel:   fieldConfig.Label,
			FieldType:    "string", // 默认字段类型
			DefaultValue: fieldConfig.DefaultPermission,
			Description:  fieldConfig.Description,
			IsActive:     true,
		}

		if err := m.db.Create(&fieldPerm).Error; err != nil {
			return fmt.Errorf("创建字段权限失败: %v", err)
		}

		log.Printf("创建字段权限: %s.%s -> %s", tableName, fieldConfig.FieldName, fieldConfig.DefaultPermission)
	} else if result.Error != nil {
		return fmt.Errorf("查询字段权限失败: %v", result.Error)
	} else {
		// 字段权限已存在，更新配置（保持is_active状态不变）
		updates := map[string]interface{}{
			"field_label":    fieldConfig.Label,
			"default_value":  fieldConfig.DefaultPermission,
			"description":    fieldConfig.Description,
		}

		if err := m.db.Model(&existingPerm).Updates(updates).Error; err != nil {
			return fmt.Errorf("更新字段权限失败: %v", err)
		}

		log.Printf("更新字段权限: %s.%s -> %s", tableName, fieldConfig.FieldName, fieldConfig.DefaultPermission)
	}

	return nil
}

// validateFieldPermissionConfigs 验证字段权限配置
func validateFieldPermissionConfigs(configs []TableFieldConfig) error {
	validPermissions := map[string]bool{
		"default":  true,
		"readonly": true,
		"hidden":   true,
	}

	tableFieldMap := make(map[string]map[string]bool)

	for _, tableConfig := range configs {
		if tableConfig.TableName == "" {
			return fmt.Errorf("表名不能为空")
		}

		if tableFieldMap[tableConfig.TableName] == nil {
			tableFieldMap[tableConfig.TableName] = make(map[string]bool)
		}

		for _, fieldConfig := range tableConfig.Fields {
			if fieldConfig.FieldName == "" {
				return fmt.Errorf("表 %s 中存在空字段名", tableConfig.TableName)
			}

			if fieldConfig.Label == "" {
				return fmt.Errorf("表 %s 字段 %s 的标签不能为空", tableConfig.TableName, fieldConfig.FieldName)
			}

			if !validPermissions[fieldConfig.DefaultPermission] {
				return fmt.Errorf("表 %s 字段 %s 的默认权限 %s 无效，只支持: default, readonly, hidden", 
					tableConfig.TableName, fieldConfig.FieldName, fieldConfig.DefaultPermission)
			}

			// 检查字段名重复
			if tableFieldMap[tableConfig.TableName][fieldConfig.FieldName] {
				return fmt.Errorf("表 %s 中字段 %s 重复定义", tableConfig.TableName, fieldConfig.FieldName)
			}
			tableFieldMap[tableConfig.TableName][fieldConfig.FieldName] = true
		}
	}

	log.Printf("字段权限配置验证通过，共 %d 个表", len(configs))
	return nil
}