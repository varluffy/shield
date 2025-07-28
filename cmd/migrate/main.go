// Package main provides database migration and management utility.
// It handles database schema creation, migration operations, and user management.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/varluffy/shield/internal/config"
	"github.com/varluffy/shield/internal/database"
	"github.com/varluffy/shield/pkg/logger"
)

func main() {
	// 解析命令行参数
	var configPath string
	var action string

	// 迁移相关参数
	var clean bool
	var addUUIDs bool
	var verifyOnly bool
	var initPermissions bool

	// 用户管理相关参数
	var email string
	var password string
	var name string
	var roleCode string
	var tenantID string

	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.StringVar(&action, "action", "migrate", "Action: migrate, create-user, update-user, list-users, create-test-users, clean-test-users, list-test-users")

	// 迁移参数
	flag.BoolVar(&clean, "clean", false, "Clean all tables before migration")
	flag.BoolVar(&addUUIDs, "add-uuids", false, "Add UUID columns to existing tables and generate UUIDs")
	flag.BoolVar(&verifyOnly, "verify", false, "Only verify existing UUID data")
	flag.BoolVar(&initPermissions, "init-permissions", false, "Initialize system permissions and roles")

	// 用户管理参数
	flag.StringVar(&email, "email", "", "User email")
	flag.StringVar(&password, "password", "", "User password")
	flag.StringVar(&name, "name", "", "User name")
	flag.StringVar(&roleCode, "role", "system_admin", "Role code (system_admin, tenant_admin)")
	flag.StringVar(&tenantID, "tenant", "", "Tenant ID (optional, uses default tenant if not specified)")

	flag.Parse()

	// 加载配置
	loader := config.NewConfigLoader()
	cfg, err := loader.LoadConfig(configPath)
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

	// 根据不同的 action 执行相应操作
	switch action {
	case "migrate":
		// 如果需要清理，先删除所有表
		if clean {
			if err := database.CleanDatabase(db); err != nil {
				log.Fatalf("Failed to clean database: %v", err)
			}
		}

		// 创建迁移实例
		migration := NewMigration(db, appLogger.Logger)

		// 处理UUID相关操作
		if verifyOnly {
			// 只验证UUID数据
			log.Println("🔍 Verifying UUID data...")
			if err := migration.VerifyMigration(); err != nil {
				log.Fatalf("❌ UUID verification failed: %v", err)
			}
			log.Println("✅ UUID verification completed successfully")
			return
		}

		if addUUIDs {
			// 执行UUID迁移
			log.Println("🚀 Starting UUID migration...")
			if err := migration.RunMigration(); err != nil {
				log.Fatalf("❌ UUID migration failed: %v", err)
			}

			// 验证迁移结果
			if err := migration.VerifyMigration(); err != nil {
				log.Fatalf("❌ UUID migration verification failed: %v", err)
			}

			log.Println("🎉 UUID migration completed successfully!")
			return
		}

		if initPermissions {
			// 执行权限系统初始化
			log.Println("🚀 Starting permission system initialization...")
			if err := migration.RunPermissionMigration(); err != nil {
				log.Fatalf("❌ Permission system initialization failed: %v", err)
			}
			log.Println("🎉 Permission system initialization completed successfully!")
			return
		}

		// 执行标准数据库迁移
		if err := database.MigrateDatabase(db); err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}

		log.Println("Database migration completed successfully")

	case "create-user":
		if email == "" || password == "" || name == "" {
			fmt.Println("Usage: -action=create-user -email=admin@example.com -password=admin123 -name=Admin [-role=system_admin] [-tenant=tenant_id]")
			os.Exit(1)
		}
		if err := createAdmin(db, email, password, name, roleCode, tenantID); err != nil {
			log.Fatalf("Failed to create admin: %v", err)
		}

	case "update-user":
		if email == "" {
			fmt.Println("Usage: -action=update-user -email=admin@example.com [-password=newpass] [-name=NewName] [-role=system_admin] [-tenant=tenant_id]")
			os.Exit(1)
		}
		if err := updateAdmin(db, email, password, name, roleCode, tenantID); err != nil {
			log.Fatalf("Failed to update admin: %v", err)
		}

	case "list-users":
		if err := listAdmins(db); err != nil {
			log.Fatalf("Failed to list admins: %v", err)
		}

	case "create-test-users":
		if err := CreateStandardTestUsers(db); err != nil {
			log.Fatalf("Failed to create test users: %v", err)
		}

	case "clean-test-users":
		if err := CleanTestUsers(db); err != nil {
			log.Fatalf("Failed to clean test users: %v", err)
		}

	case "list-test-users":
		if err := ListTestUsers(db); err != nil {
			log.Fatalf("Failed to list test users: %v", err)
		}

	default:
		fmt.Printf("Unknown action: %s\n", action)
		fmt.Println("Available actions: migrate, create-user, update-user, list-users, create-test-users, clean-test-users, list-test-users")
		os.Exit(1)
	}
}
