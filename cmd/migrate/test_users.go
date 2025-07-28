package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/varluffy/shield/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TestUser 测试用户配置
type TestUser struct {
	Email     string
	Password  string
	Name      string
	TenantID  uint64
	RoleCode  string
	IsActive  bool
}

// GetStandardTestUsers 获取标准测试用户配置
func GetStandardTestUsers() []TestUser {
	return []TestUser{
		{
			Email:    "admin@system.test",
			Password: "admin123",
			Name:     "系统管理员",
			TenantID: 0, // 系统租户
			RoleCode: "system_admin",
			IsActive: true,
		},
		{
			Email:    "admin@tenant.test",
			Password: "admin123",
			Name:     "租户管理员",
			TenantID: 1, // 默认租户
			RoleCode: "tenant_admin",
			IsActive: true,
		},
		{
			Email:    "user@tenant.test",
			Password: "user123",
			Name:     "普通用户",
			TenantID: 1, // 默认租户
			RoleCode: "user",
			IsActive: true,
		},
		{
			Email:    "test@example.com",
			Password: "test123",
			Name:     "测试用户",
			TenantID: 1, // 默认租户
			RoleCode: "user",
			IsActive: true,
		},
	}
}

// CreateStandardTestUsers 创建标准测试用户
func CreateStandardTestUsers(db *gorm.DB) error {
	testUsers := GetStandardTestUsers()

	fmt.Println("🚀 开始创建标准测试用户...")

	for _, testUser := range testUsers {
		if err := createTestUser(db, testUser); err != nil {
			log.Printf("❌ 创建用户 %s 失败: %v", testUser.Email, err)
			continue
		}
		fmt.Printf("✅ 成功创建测试用户: %s (密码: %s)\n", testUser.Email, testUser.Password)
	}

	fmt.Println("🎉 标准测试用户创建完成!")
	return nil
}

// createTestUser 创建单个测试用户
func createTestUser(db *gorm.DB, testUser TestUser) error {
	// 检查用户是否已存在
	var existingUser models.User
	err := db.Where("tenant_id = ? AND email = ?", testUser.TenantID, testUser.Email).First(&existingUser).Error
	if err == nil {
		// 用户已存在，更新密码
		return updateTestUserPassword(db, &existingUser, testUser.Password)
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("检查用户是否存在时出错: %w", err)
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户
	user := models.User{
		TenantModel: models.TenantModel{TenantID: testUser.TenantID},
		Email:       testUser.Email,
		Password:    string(hashedPassword),
		Name:        testUser.Name,
		Status:      "active",
		Language:    "zh",
		Timezone:    "Asia/Shanghai",
	}

	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	// 分配角色（如果角色存在）
	if testUser.RoleCode != "" {
		if err := assignRoleToUser(db, user.ID, user.TenantID, testUser.RoleCode); err != nil {
			log.Printf("⚠️ 为用户 %s 分配角色 %s 失败: %v", testUser.Email, testUser.RoleCode, err)
		}
	}

	return nil
}

// updateTestUserPassword 更新测试用户密码
func updateTestUserPassword(db *gorm.DB, user *models.User, newPassword string) error {
	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新密码
	if err := db.Model(user).Update("password", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("更新密码失败: %w", err)
	}

	fmt.Printf("🔄 已更新用户 %s 的密码\n", user.Email)
	return nil
}

// assignRoleToUser 为用户分配角色
func assignRoleToUser(db *gorm.DB, userID, tenantID uint64, roleCode string) error {
	// 查找角色
	var role models.Role
	err := db.Where("tenant_id = ? AND code = ?", tenantID, roleCode).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果是系统角色，尝试在系统租户中查找
			if tenantID != 0 {
				err = db.Where("tenant_id = 0 AND code = ?", roleCode).First(&role).Error
			}
			if err != nil {
				return fmt.Errorf("角色 %s 不存在", roleCode)
			}
		} else {
			return fmt.Errorf("查找角色失败: %w", err)
		}
	}

	// 检查用户是否已有该角色
	var existingUserRole models.UserRole
	err = db.Where("user_id = ? AND role_id = ?", userID, role.ID).First(&existingUserRole).Error
	if err == nil {
		// 用户已有该角色，确保激活状态
		if !existingUserRole.IsActive {
			existingUserRole.IsActive = true
			db.Save(&existingUserRole)
		}
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("检查用户角色失败: %w", err)
	}

	// 分配角色给用户
	userRole := models.UserRole{
		UserID:   userID,
		RoleID:   role.ID,
		TenantID: tenantID,
		IsActive: true,
	}

	if err := db.Create(&userRole).Error; err != nil {
		return fmt.Errorf("分配角色失败: %w", err)
	}

	return nil
}

// CleanTestUsers 清理测试用户
func CleanTestUsers(db *gorm.DB) error {
	testUsers := GetStandardTestUsers()

	fmt.Println("🧹 开始清理测试用户...")

	for _, testUser := range testUsers {
		var user models.User
		err := db.Where("tenant_id = ? AND email = ?", testUser.TenantID, testUser.Email).First(&user).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				continue // 用户不存在，跳过
			}
			log.Printf("❌ 查找用户 %s 失败: %v", testUser.Email, err)
			continue
		}

		// 删除用户角色
		db.Where("user_id = ?", user.ID).Delete(&models.UserRole{})

		// 删除用户
		if err := db.Delete(&user).Error; err != nil {
			log.Printf("❌ 删除用户 %s 失败: %v", testUser.Email, err)
			continue
		}

		fmt.Printf("🗑️ 已删除测试用户: %s\n", testUser.Email)
	}

	fmt.Println("🎉 测试用户清理完成!")
	return nil
}

// ListTestUsers 列出所有测试用户
func ListTestUsers(db *gorm.DB) error {
	testUsers := GetStandardTestUsers()

	fmt.Println("📋 测试用户列表:")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("%-25s %-15s %-20s %-10s %-15s\n", "邮箱", "密码", "姓名", "租户ID", "状态")
	fmt.Println(strings.Repeat("-", 80))

	for _, testUser := range testUsers {
		var user models.User
		err := db.Where("tenant_id = ? AND email = ?", testUser.TenantID, testUser.Email).First(&user).Error

		status := "不存在"
		if err == nil {
			status = user.Status
		}

		fmt.Printf("%-25s %-15s %-20s %-10d %-15s\n",
			testUser.Email,
			testUser.Password,
			testUser.Name,
			testUser.TenantID,
			status)
	}

	fmt.Println(strings.Repeat("=", 80))
	return nil
}