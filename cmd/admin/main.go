// Package main provides admin user management utility.
// It handles admin user creation and management operations.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/varluffy/shield/internal/config"
	"github.com/varluffy/shield/internal/database"
	"github.com/varluffy/shield/internal/models"
	"github.com/varluffy/shield/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	var (
		configPath = flag.String("config", "", "Path to config file")
		email      = flag.String("email", "", "Admin email")
		password   = flag.String("password", "", "Admin password")
		name       = flag.String("name", "", "Admin name")
		roleCode   = flag.String("role", "system_admin", "Role code (system_admin, tenant_admin)")
		tenantID   = flag.String("tenant", "", "Tenant ID (optional, uses default tenant if not specified)")
		action     = flag.String("action", "create", "Action: create, update, list")
	)
	flag.Parse()

	// 加载配置
	loader := config.NewConfigLoader()
	cfg, err := loader.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建日志器
	logConfig := &logger.LogConfig{
		Level:      cfg.Log.Level,
		Format:     cfg.Log.Format,
		Output:     cfg.Log.Output,
		MaxSize:    cfg.Log.MaxSize,
		MaxAge:     cfg.Log.MaxAge,
		MaxBackups: cfg.Log.MaxBackups,
		Compress:   cfg.Log.Compress,
	}

	appLogger, err := logger.NewLoggerWithConfig(logConfig)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// 连接数据库
	db, err := database.NewMySQLConnection(cfg.Database, appLogger.Logger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 执行操作
	switch *action {
	case "create":
		if *email == "" || *password == "" || *name == "" {
			fmt.Println("Usage: -action=create -email=admin@example.com -password=admin123 -name=Admin [-role=system_admin] [-tenant=tenant_id]")
			os.Exit(1)
		}
		if err := createAdmin(db, *email, *password, *name, *roleCode, *tenantID); err != nil {
			log.Fatalf("Failed to create admin: %v", err)
		}
	case "update":
		if *email == "" {
			fmt.Println("Usage: -action=update -email=admin@example.com [-password=newpass] [-name=NewName] [-role=system_admin] [-tenant=tenant_id]")
			os.Exit(1)
		}
		if err := updateAdmin(db, *email, *password, *name, *roleCode, *tenantID); err != nil {
			log.Fatalf("Failed to update admin: %v", err)
		}
	case "list":
		if err := listAdmins(db); err != nil {
			log.Fatalf("Failed to list admins: %v", err)
		}
	default:
		fmt.Println("Available actions: create, update, list")
		os.Exit(1)
	}
}

func createAdmin(db *gorm.DB, email, password, name, roleCode, tenantID string) error {
	// 获取默认租户
	var tenant models.Tenant
	var tenantIDUint64 uint64
	
	if tenantID == "" {
		if err := db.Where("domain = ?", "default.ultrafit.com").First(&tenant).Error; err != nil {
			return fmt.Errorf("failed to find default tenant: %w", err)
		}
		tenantIDUint64 = tenant.ID
	} else {
		// 尝试将字符串转换为uint64
		parsedID, err := strconv.ParseUint(tenantID, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid tenant ID format: %w", err)
		}
		tenantIDUint64 = parsedID
		
		if err := db.Where("id = ?", tenantIDUint64).First(&tenant).Error; err != nil {
			return fmt.Errorf("failed to find tenant: %w", err)
		}
	}

	// 检查用户是否已存在
	var existingUser models.User
	if err := db.Where("tenant_id = ? AND email = ?", tenantIDUint64, email).First(&existingUser).Error; err == nil {
		return fmt.Errorf("user with email %s already exists in tenant %s", email, tenant.Name)
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := models.User{
		TenantModel: models.TenantModel{TenantID: tenantIDUint64},
		Email:       email,
		Password:    string(hashedPassword),
		Name:        name,
		Status:      "active",
		Language:    "zh",
		Timezone:    "Asia/Shanghai",
	}

	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// 查找角色
	var role models.Role
	if err := db.Where("tenant_id = ? AND code = ?", tenantIDUint64, roleCode).First(&role).Error; err != nil {
		return fmt.Errorf("failed to find role %s: %w", roleCode, err)
	}

	// 分配角色给用户
	userRole := models.UserRole{
		UserID:   user.ID,
		RoleID:   role.ID,
		TenantID: tenantIDUint64,
		IsActive: true,
	}

	if err := db.Create(&userRole).Error; err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	fmt.Printf("Admin user created successfully:\n")
	fmt.Printf("- ID: %d\n", user.ID)
	fmt.Printf("- Email: %s\n", user.Email)
	fmt.Printf("- Name: %s\n", user.Name)
	fmt.Printf("- Role: %s (%s)\n", role.Name, role.Code)
	fmt.Printf("- Tenant: %s (%d)\n", tenant.Name, tenant.ID)
	fmt.Printf("- Status: %s\n", user.Status)

	return nil
}

func updateAdmin(db *gorm.DB, email, password, name, roleCode, tenantID string) error {
	// 查找用户
	var user models.User
	query := db.Where("email = ?", email)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	
	if err := query.First(&user).Error; err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// 更新用户信息
	updates := make(map[string]interface{})
	
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		updates["password"] = string(hashedPassword)
	}
	
	if name != "" {
		updates["name"] = name
	}

	if len(updates) > 0 {
		if err := db.Model(&user).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		fmt.Printf("User updated successfully\n")
	}

	// 更新角色
	if roleCode != "" {
		// 查找新角色
		var newRole models.Role
		if err := db.Where("tenant_id = ? AND code = ?", user.TenantID, roleCode).First(&newRole).Error; err != nil {
			return fmt.Errorf("failed to find role %s: %w", roleCode, err)
		}

		// 删除现有角色
		if err := db.Where("user_id = ?", user.ID).Delete(&models.UserRole{}).Error; err != nil {
			return fmt.Errorf("failed to remove existing roles: %w", err)
		}

		// 分配新角色
		userRole := models.UserRole{
			UserID:   user.ID,
			RoleID:   newRole.ID,
			TenantID: user.TenantID,
			IsActive: true,
		}

		if err := db.Create(&userRole).Error; err != nil {
			return fmt.Errorf("failed to assign new role: %w", err)
		}

		fmt.Printf("Role updated to: %s (%s)\n", newRole.Name, newRole.Code)
	}

	return nil
}

func listAdmins(db *gorm.DB) error {
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	fmt.Printf("Admin Users:\n")
	fmt.Printf("%-8s %-36s %-30s %-20s %-15s %-20s %-10s\n", "ID", "UUID", "Email", "Name", "Status", "Tenant", "Roles")
	fmt.Printf("%s\n", strings.Repeat("-", 150))

	for _, user := range users {
		// 获取用户的租户信息
		var tenant models.Tenant
		tenantName := "Unknown"
		if err := db.Where("id = ?", user.TenantID).First(&tenant).Error; err == nil {
			tenantName = tenant.Name
		}

		// 获取用户的角色信息
		var userRoles []models.UserRole
		var roleNames []string
		if err := db.Where("user_id = ?", user.ID).Find(&userRoles).Error; err == nil {
			for _, ur := range userRoles {
				var role models.Role
				if err := db.Where("id = ?", ur.RoleID).First(&role).Error; err == nil {
					roleNames = append(roleNames, role.Name)
				}
			}
		}

		fmt.Printf("%-8d %-36s %-30s %-20s %-15s %-20s %-10s\n", 
			user.ID,
			user.UUID, 
			user.Email, 
			user.Name, 
			user.Status, 
			tenantName,
			strings.Join(roleNames, ","))
	}

	return nil
} 